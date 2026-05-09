package codex

import (
	"fmt"

	"github.com/openai/codex/sdk/go/protocol"
)

type InputItem interface {
	wire() protocol.UserInput
}

type TextInput struct {
	Text string
}

func (i TextInput) wire() protocol.UserInput {
	return protocol.UserInput{Type: "text", Text: i.Text}
}

type ImageInput struct {
	URL string
}

func (i ImageInput) wire() protocol.UserInput {
	return protocol.UserInput{Type: "image", URL: i.URL}
}

type LocalImageInput struct {
	Path string
}

func (i LocalImageInput) wire() protocol.UserInput {
	return protocol.UserInput{Type: "localImage", Path: i.Path}
}

type SkillInput struct {
	Name string
	Path string
}

func (i SkillInput) wire() protocol.UserInput {
	return protocol.UserInput{Type: "skill", Name: i.Name, Path: i.Path}
}

type MentionInput struct {
	Name string
	Path string
}

func (i MentionInput) wire() protocol.UserInput {
	return protocol.UserInput{Type: "mention", Name: i.Name, Path: i.Path}
}

func normalizeRunInput(input any) ([]protocol.UserInput, error) {
	switch value := input.(type) {
	case string:
		return []protocol.UserInput{{Type: "text", Text: value}}, nil
	case InputItem:
		return []protocol.UserInput{value.wire()}, nil
	case []InputItem:
		items := make([]protocol.UserInput, 0, len(value))
		for _, item := range value {
			items = append(items, item.wire())
		}
		return items, nil
	default:
		return nil, fmt.Errorf("unsupported input type %T", input)
	}
}
