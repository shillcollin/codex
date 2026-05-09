package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	irpc "github.com/openai/codex/sdk/go/internal/jsonrpc"
	"github.com/openai/codex/sdk/go/internal/transport/stdio"
	"github.com/openai/codex/sdk/go/protocol"
)

type Client struct {
	transport       *stdio.Transport
	approvalHandler ApprovalHandler

	requestID atomic.Uint64

	pendingMu sync.Mutex
	pending   map[string]chan responseResult

	subsMu     sync.Mutex
	clientSubs []chan protocol.Notification
	threadSubs map[string][]chan protocol.Notification
	turnSubs   map[string][]chan protocol.Notification

	meta protocol.InitializeResponse

	closeOnce sync.Once
	closed    chan struct{}
	closeMu   sync.Mutex
	closeErr  error
}

type responseResult struct {
	result json.RawMessage
	err    error
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	client := &Client{
		approvalHandler: effectiveApprovalHandler(cfg.ApprovalHandler),
		pending:         make(map[string]chan responseResult),
		threadSubs:      make(map[string][]chan protocol.Notification),
		turnSubs:        make(map[string][]chan protocol.Notification),
		closed:          make(chan struct{}),
	}

	transport, err := stdio.Start(ctx, stdio.Config{
		CodexBin:        cfg.CodexBin,
		ConfigOverrides: cfg.ConfigOverrides,
		Cwd:             cfg.Cwd,
		Env:             cfg.Env,
	})
	if err != nil {
		return nil, err
	}
	client.transport = transport
	go client.readLoop()

	params := protocol.InitializeParams{
		ClientInfo: protocol.ClientInfo{
			Name:    defaultString(cfg.ClientName, "codex_go_sdk"),
			Title:   defaultString(cfg.ClientTitle, "Codex Go SDK"),
			Version: defaultString(cfg.ClientVersion, "0.1.0"),
		},
		Capabilities: &protocol.InitializeCapabilities{
			ExperimentalAPI: cfg.ExperimentalAPI,
		},
	}

	var initResp protocol.InitializeResponse
	if err := client.request(ctx, "initialize", params, &initResp); err != nil {
		_ = client.Close()
		return nil, err
	}
	if err := validateInitialize(&initResp); err != nil {
		_ = client.Close()
		return nil, err
	}
	client.meta = initResp

	if err := client.notify("initialized", map[string]any{}); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func (c *Client) Close() error {
	c.shutdown(nil)
	return nil
}

func (c *Client) Metadata() protocol.InitializeResponse {
	return c.meta
}

func (c *Client) Notifications(ctx context.Context) (<-chan protocol.Notification, <-chan error) {
	source, unsubscribe := c.subscribeClient()
	return c.forwardNotifications(ctx, source, unsubscribe, nil)
}

func (c *Client) Models(ctx context.Context, includeHidden bool) (*protocol.ModelListResponse, error) {
	var resp protocol.ModelListResponse
	err := c.request(ctx, "model/list", map[string]any{"includeHidden": includeHidden}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ThreadStart(ctx context.Context, params protocol.ThreadStartParams) (*Thread, error) {
	var resp protocol.ThreadStartResponse
	if err := c.request(ctx, "thread/start", params, &resp); err != nil {
		return nil, err
	}
	return &Thread{client: c, ID: resp.Thread.ID}, nil
}

func (c *Client) ThreadResume(ctx context.Context, threadID string, params protocol.ThreadResumeParams) (*Thread, error) {
	payload := struct {
		ThreadID string `json:"threadId"`
		protocol.ThreadResumeParams
	}{
		ThreadID:           threadID,
		ThreadResumeParams: params,
	}
	var resp protocol.ThreadResumeResponse
	if err := c.request(ctx, "thread/resume", payload, &resp); err != nil {
		return nil, err
	}
	return &Thread{client: c, ID: resp.Thread.ID}, nil
}

func (c *Client) ThreadFork(ctx context.Context, threadID string, params protocol.ThreadForkParams) (*Thread, error) {
	payload := struct {
		ThreadID string `json:"threadId"`
		protocol.ThreadForkParams
	}{
		ThreadID:         threadID,
		ThreadForkParams: params,
	}
	var resp protocol.ThreadForkResponse
	if err := c.request(ctx, "thread/fork", payload, &resp); err != nil {
		return nil, err
	}
	return &Thread{client: c, ID: resp.Thread.ID}, nil
}

func (c *Client) ThreadList(ctx context.Context, params protocol.ThreadListParams) (*protocol.ThreadListResponse, error) {
	var resp protocol.ThreadListResponse
	if err := c.request(ctx, "thread/list", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ThreadArchive(ctx context.Context, threadID string) error {
	return c.request(ctx, "thread/archive", map[string]any{"threadId": threadID}, &protocol.ThreadArchiveResponse{})
}

func (c *Client) ThreadUnarchive(ctx context.Context, threadID string) (*Thread, error) {
	var resp protocol.ThreadUnarchiveResponse
	if err := c.request(ctx, "thread/unarchive", map[string]any{"threadId": threadID}, &resp); err != nil {
		return nil, err
	}
	return &Thread{client: c, ID: resp.Thread.ID}, nil
}

func (c *Client) request(ctx context.Context, method string, params any, out any) error {
	id := irpc.NewStringRequestID(fmt.Sprintf("%d", c.requestID.Add(1)))
	waiter := make(chan responseResult, 1)
	key := id.Key()

	c.pendingMu.Lock()
	c.pending[key] = waiter
	c.pendingMu.Unlock()

	cleanup := func() {
		c.pendingMu.Lock()
		delete(c.pending, key)
		c.pendingMu.Unlock()
	}

	if err := c.transport.WriteJSON(irpc.Request{ID: id, Method: method, Params: params}); err != nil {
		cleanup()
		return err
	}

	select {
	case result := <-waiter:
		cleanup()
		if result.err != nil {
			return result.err
		}
		if out == nil || len(result.result) == 0 {
			return nil
		}
		return json.Unmarshal(result.result, out)
	case <-ctx.Done():
		cleanup()
		return ctx.Err()
	case <-c.closed:
		cleanup()
		return c.closedErr()
	}
}

func (c *Client) notify(method string, params any) error {
	return c.transport.WriteJSON(irpc.Notification{Method: method, Params: params})
}

func (c *Client) readLoop() {
	for {
		line, err := c.transport.ReadLine()
		if err != nil {
			c.shutdown(err)
			return
		}

		env, err := irpc.ParseEnvelope(line)
		if err != nil {
			c.shutdown(fmt.Errorf("decode json-rpc envelope: %w", err))
			return
		}

		switch {
		case env.Method != "" && env.ID != nil:
			c.handleServerRequest(*env.ID, env.Method, env.Params)
		case env.Method != "":
			c.dispatch(protocol.Notification{Method: env.Method, Params: env.Params})
		case env.ID != nil:
			c.deliverResponse(*env.ID, env.Result, env.Error)
		}
	}
}

func (c *Client) handleServerRequest(id irpc.RequestID, method string, params json.RawMessage) {
	result, err := c.approvalHandler(context.Background(), method, params)
	if err != nil {
		_ = c.transport.WriteJSON(map[string]any{
			"id": id,
			"error": map[string]any{
				"code":    -32000,
				"message": err.Error(),
			},
		})
		return
	}
	if result == nil {
		result = map[string]any{}
	}
	_ = c.transport.WriteJSON(map[string]any{"id": id, "result": result})
}

func (c *Client) deliverResponse(id irpc.RequestID, result json.RawMessage, errBody *irpc.ErrorBody) {
	c.pendingMu.Lock()
	waiter := c.pending[id.Key()]
	c.pendingMu.Unlock()
	if waiter == nil {
		return
	}
	if errBody != nil {
		waiter <- responseResult{
			err: &irpc.Error{
				Code:    errBody.Code,
				Message: errBody.Message,
				Data:    errBody.Data,
			},
		}
		return
	}
	waiter <- responseResult{result: result}
}

func (c *Client) dispatch(event protocol.Notification) {
	c.subsMu.Lock()
	clientSubscribers := append([]chan protocol.Notification(nil), c.clientSubs...)
	threadSubscribers := append([]chan protocol.Notification(nil), c.threadSubs[event.ThreadID()]...)
	turnSubscribers := append([]chan protocol.Notification(nil), c.turnSubs[event.TurnID()]...)
	c.subsMu.Unlock()

	for _, ch := range clientSubscribers {
		select {
		case ch <- event:
		default:
		}
	}
	for _, ch := range threadSubscribers {
		select {
		case ch <- event:
		default:
		}
	}
	for _, ch := range turnSubscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (c *Client) subscribeClient() (<-chan protocol.Notification, func()) {
	ch := make(chan protocol.Notification, 32)
	c.subsMu.Lock()
	c.clientSubs = append(c.clientSubs, ch)
	c.subsMu.Unlock()

	cancel := func() {
		c.subsMu.Lock()
		defer c.subsMu.Unlock()
		filtered := c.clientSubs[:0]
		for _, existing := range c.clientSubs {
			if existing != ch {
				filtered = append(filtered, existing)
			}
		}
		c.clientSubs = filtered
		close(ch)
	}

	return ch, cancel
}

func (c *Client) subscribeThread(threadID string) (<-chan protocol.Notification, func()) {
	ch := make(chan protocol.Notification, 32)
	c.subsMu.Lock()
	c.threadSubs[threadID] = append(c.threadSubs[threadID], ch)
	c.subsMu.Unlock()

	cancel := func() {
		c.subsMu.Lock()
		defer c.subsMu.Unlock()
		subs := c.threadSubs[threadID]
		filtered := subs[:0]
		for _, existing := range subs {
			if existing != ch {
				filtered = append(filtered, existing)
			}
		}
		if len(filtered) == 0 {
			delete(c.threadSubs, threadID)
		} else {
			c.threadSubs[threadID] = filtered
		}
		close(ch)
	}

	return ch, cancel
}

func (c *Client) subscribeTurn(turnID string) (<-chan protocol.Notification, func()) {
	ch := make(chan protocol.Notification, 32)
	c.subsMu.Lock()
	c.turnSubs[turnID] = append(c.turnSubs[turnID], ch)
	c.subsMu.Unlock()

	cancel := func() {
		c.subsMu.Lock()
		defer c.subsMu.Unlock()
		subs := c.turnSubs[turnID]
		filtered := subs[:0]
		for _, existing := range subs {
			if existing != ch {
				filtered = append(filtered, existing)
			}
		}
		if len(filtered) == 0 {
			delete(c.turnSubs, turnID)
		} else {
			c.turnSubs[turnID] = filtered
		}
		close(ch)
	}

	return ch, cancel
}

func (c *Client) forwardNotifications(
	ctx context.Context,
	source <-chan protocol.Notification,
	unsubscribe func(),
	stop func(protocol.Notification) bool,
) (<-chan protocol.Notification, <-chan error) {
	out := make(chan protocol.Notification, 32)
	errCh := make(chan error, 1)

	go func() {
		defer unsubscribe()
		defer close(out)
		defer close(errCh)

		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-c.closed:
				if err := c.closedErr(); err != nil {
					errCh <- err
				}
				return
			case event, ok := <-source:
				if !ok {
					return
				}
				out <- event
				if stop != nil && stop(event) {
					return
				}
			}
		}
	}()

	return out, errCh
}

func (c *Client) shutdown(err error) {
	c.closeOnce.Do(func() {
		c.closeMu.Lock()
		c.closeErr = err
		c.closeMu.Unlock()

		_ = c.transport.Close()

		c.pendingMu.Lock()
		for id, waiter := range c.pending {
			waiter <- responseResult{err: coalesceErr(err, errors.New("client closed"))}
			close(waiter)
			delete(c.pending, id)
		}
		c.pendingMu.Unlock()

		c.subsMu.Lock()
		for _, sub := range c.clientSubs {
			close(sub)
		}
		c.clientSubs = nil
		for turnID, subs := range c.turnSubs {
			for _, sub := range subs {
				close(sub)
			}
			delete(c.turnSubs, turnID)
		}
		for threadID, subs := range c.threadSubs {
			for _, sub := range subs {
				close(sub)
			}
			delete(c.threadSubs, threadID)
		}
		c.subsMu.Unlock()

		close(c.closed)
	})
}

func (c *Client) closedErr() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	return coalesceErr(c.closeErr, errors.New("client closed"))
}

type Thread struct {
	client *Client
	ID     string
}

func (t *Thread) Notifications(ctx context.Context) (<-chan protocol.Notification, <-chan error) {
	source, unsubscribe := t.client.subscribeThread(t.ID)
	return t.client.forwardNotifications(ctx, source, unsubscribe, nil)
}

func (t *Thread) Run(ctx context.Context, input any, opts RunOptions) (*RunResult, error) {
	handle, err := t.Turn(ctx, input, opts)
	if err != nil {
		return nil, err
	}
	collected, err := handle.collect(ctx)
	if err != nil {
		return nil, err
	}
	return collected.run, nil
}

func (t *Thread) Turn(ctx context.Context, input any, opts RunOptions) (*TurnHandle, error) {
	wireInput, err := normalizeRunInput(input)
	if err != nil {
		return nil, err
	}
	params := protocol.TurnStartParams{
		ThreadID:          t.ID,
		Input:             wireInput,
		ApprovalPolicy:    opts.ApprovalPolicy,
		ApprovalsReviewer: opts.ApprovalsReviewer,
		Cwd:               opts.Cwd,
		Effort:            opts.Effort,
		Model:             opts.Model,
		OutputSchema:      opts.OutputSchema,
		Personality:       opts.Personality,
		SandboxPolicy:     opts.SandboxPolicy,
		ServiceTier:       opts.ServiceTier,
		Summary:           opts.Summary,
	}
	var resp protocol.TurnStartResponse
	if err := t.client.request(ctx, "turn/start", params, &resp); err != nil {
		return nil, err
	}
	return &TurnHandle{client: t.client, ThreadID: t.ID, TurnID: resp.Turn.ID}, nil
}

func (t *Thread) Read(ctx context.Context, includeTurns bool) (*protocol.ThreadReadResponse, error) {
	var resp protocol.ThreadReadResponse
	if err := t.client.request(ctx, "thread/read", map[string]any{
		"threadId":     t.ID,
		"includeTurns": includeTurns,
	}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t *Thread) SetName(ctx context.Context, name string) error {
	return t.client.request(ctx, "thread/name/set", map[string]any{
		"threadId": t.ID,
		"name":     name,
	}, &protocol.ThreadSetNameResponse{})
}

func (t *Thread) Compact(ctx context.Context) error {
	return t.client.request(ctx, "thread/compact/start", map[string]any{"threadId": t.ID}, &protocol.ThreadCompactStartResponse{})
}

type TurnHandle struct {
	client   *Client
	ThreadID string
	TurnID   string
}

func (h *TurnHandle) Stream(ctx context.Context) (<-chan protocol.Notification, <-chan error) {
	source, unsubscribe := h.client.subscribeTurn(h.TurnID)
	return h.client.forwardNotifications(ctx, source, unsubscribe, func(event protocol.Notification) bool {
		return event.Method == "turn/completed" && event.TurnID() == h.TurnID
	})
}

func (h *TurnHandle) Run(ctx context.Context) (*protocol.Turn, error) {
	collected, err := h.collect(ctx)
	if err != nil {
		return nil, err
	}
	return collected.turn, nil
}

func (h *TurnHandle) Steer(ctx context.Context, input any) error {
	wireInput, err := normalizeRunInput(input)
	if err != nil {
		return err
	}
	return h.client.request(ctx, "turn/steer", map[string]any{
		"threadId":       h.ThreadID,
		"expectedTurnId": h.TurnID,
		"input":          wireInput,
	}, &protocol.TurnSteerResponse{})
}

func (h *TurnHandle) Interrupt(ctx context.Context) error {
	return h.client.request(ctx, "turn/interrupt", map[string]any{
		"threadId": h.ThreadID,
		"turnId":   h.TurnID,
	}, &protocol.TurnInterruptResponse{})
}

type collectedTurn struct {
	run  *RunResult
	turn *protocol.Turn
}

func (h *TurnHandle) collect(ctx context.Context) (*collectedTurn, error) {
	events, errs := h.Stream(ctx)
	items := make([]protocol.ThreadItem, 0, 16)
	var usage *protocol.ThreadTokenUsage
	var completed *protocol.Turn

	for {
		select {
		case err, ok := <-errs:
			if ok && err != nil && !errors.Is(err, context.Canceled) {
				return nil, err
			}
			errs = nil
		case event, ok := <-events:
			if !ok {
				if completed == nil {
					return nil, errors.New("turn completed event not received")
				}
				if completed.Status == protocol.TurnStatusFailed {
					if completed.Error != nil && completed.Error.Message != "" {
						return nil, errors.New(completed.Error.Message)
					}
					return nil, fmt.Errorf("turn failed with status %s", completed.Status)
				}
				return &collectedTurn{
					run: &RunResult{
						FinalResponse: finalAssistantResponse(items),
						Items:         items,
						Usage:         usage,
					},
					turn: completed,
				}, nil
			}

			switch event.Method {
			case "item/completed":
				var payload protocol.ItemCompletedNotification
				if err := event.Decode(&payload); err == nil && payload.TurnID == h.TurnID {
					items = append(items, payload.Item)
				}
			case "thread/tokenUsage/updated":
				var payload protocol.ThreadTokenUsageUpdatedNotification
				if err := event.Decode(&payload); err == nil && payload.TurnID == h.TurnID {
					usage = &payload.TokenUsage
				}
			case "turn/completed":
				var payload protocol.TurnCompletedNotification
				if err := event.Decode(&payload); err == nil && payload.Turn.ID == h.TurnID {
					turn := payload.Turn
					if len(turn.Items) == 0 && len(items) > 0 {
						turn.Items = append([]protocol.ThreadItem(nil), items...)
					}
					completed = &turn
				}
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func finalAssistantResponse(items []protocol.ThreadItem) string {
	var unknownPhase string
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if item.Type != "agentMessage" {
			continue
		}
		if item.Phase != nil && *item.Phase == protocol.MessagePhaseFinalAnswer {
			return item.Text
		}
		if item.Phase == nil && unknownPhase == "" {
			unknownPhase = item.Text
		}
	}
	return unknownPhase
}

func effectiveApprovalHandler(handler ApprovalHandler) ApprovalHandler {
	if handler != nil {
		return handler
	}
	return defaultApprovalHandler
}

func validateInitialize(resp *protocol.InitializeResponse) error {
	userAgent := strings.TrimSpace(resp.UserAgent)
	if resp.ServerInfo == nil && userAgent != "" {
		parts := strings.SplitN(userAgent, "/", 2)
		if len(parts) == 2 {
			resp.ServerInfo = &protocol.ServerInfo{Name: parts[0], Version: parts[1]}
		}
	}
	if strings.TrimSpace(resp.UserAgent) == "" || resp.ServerInfo == nil || strings.TrimSpace(resp.ServerInfo.Name) == "" || strings.TrimSpace(resp.ServerInfo.Version) == "" {
		return fmt.Errorf("initialize response missing required metadata (user_agent=%q)", resp.UserAgent)
	}
	return nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func coalesceErr(err, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}
