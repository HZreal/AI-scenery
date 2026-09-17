package chat

import "strings"

type Mode string

const (
	ModeText Mode = "text"
	ModeJSON Mode = "json"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Messages []Message
	Mode     Mode
}

type Result struct {
	Value        any
	Model        string
	InputTokens  *int32
	OutputTokens *int32
}

type Error struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *Error) Error() string { return e.Message }

func NewError(code, message string, statusCode int) *Error {
	return &Error{Code: code, Message: message, StatusCode: statusCode}
}

func (r Request) InputCharacters() int {
	total := 0
	for _, message := range r.Messages {
		total += len(message.Content)
	}
	return total
}

func NewRequest(messages []Message, shorthand string, mode string) (Request, error) {
	if len(messages) > 0 && shorthand != "" {
		return Request{}, NewError("ambiguous_input", "messages 和 message 不能同时提供", 400)
	}
	if len(messages) == 0 && shorthand != "" {
		messages = []Message{{Role: "user", Content: shorthand}}
	}
	if len(messages) == 0 {
		return Request{}, NewError("missing_messages", "messages 或 message 必须提供", 400)
	}
	for _, message := range messages {
		if message.Role != "system" && message.Role != "user" && message.Role != "assistant" {
			return Request{}, NewError("invalid_role", "role 必须是 system、user 或 assistant", 400)
		}
		if strings.TrimSpace(message.Content) == "" {
			return Request{}, NewError("empty_message", "消息内容不能为空", 400)
		}
	}
	if mode == "" {
		mode = string(ModeText)
	}
	if mode != string(ModeText) && mode != string(ModeJSON) {
		return Request{}, NewError("invalid_mode", "mode 必须是 text 或 json", 400)
	}
	return Request{Messages: messages, Mode: Mode(mode)}, nil
}
