package jsonrpc

import (
	"encoding/json"
	"fmt"
)

type Request struct {
	ID     RequestID `json:"id"`
	Method string    `json:"method"`
	Params any       `json:"params,omitempty"`
}

type Notification struct {
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

type ErrorBody struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type Response struct {
	ID     RequestID       `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ErrorBody      `json:"error,omitempty"`
}

type Envelope struct {
	ID     *RequestID      `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ErrorBody      `json:"error,omitempty"`
}

type RequestID struct {
	stringValue  *string
	integerValue *int64
}

func NewStringRequestID(value string) RequestID {
	return RequestID{stringValue: &value}
}

func NewIntegerRequestID(value int64) RequestID {
	return RequestID{integerValue: &value}
}

func (id RequestID) String() string {
	if id.stringValue != nil {
		return *id.stringValue
	}
	if id.integerValue != nil {
		return fmt.Sprintf("%d", *id.integerValue)
	}
	return ""
}

func (id RequestID) Key() string {
	if id.stringValue != nil {
		return "s:" + *id.stringValue
	}
	if id.integerValue != nil {
		return fmt.Sprintf("i:%d", *id.integerValue)
	}
	return ""
}

func (id RequestID) MarshalJSON() ([]byte, error) {
	if id.stringValue != nil {
		return json.Marshal(*id.stringValue)
	}
	if id.integerValue != nil {
		return json.Marshal(*id.integerValue)
	}
	return []byte("null"), nil
}

func (id *RequestID) UnmarshalJSON(data []byte) error {
	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		*id = NewStringRequestID(stringValue)
		return nil
	}

	var integerValue int64
	if err := json.Unmarshal(data, &integerValue); err == nil {
		*id = NewIntegerRequestID(integerValue)
		return nil
	}

	return fmt.Errorf("unsupported request id: %s", string(data))
}

type Error struct {
	Code    int64
	Message string
	Data    json.RawMessage
}

func (e *Error) Error() string {
	return fmt.Sprintf("json-rpc error %d: %s", e.Code, e.Message)
}

func ParseEnvelope(line []byte) (Envelope, error) {
	var env Envelope
	err := json.Unmarshal(line, &env)
	return env, err
}
