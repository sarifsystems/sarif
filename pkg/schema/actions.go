package schema

type Action struct {
	*Thing
	Reply   string      `json:"reply,omitempty"`
	Name    string      `json:"name,omitempty"`
	Payload interface{} `json:"p,omitempty"`
}

type MultipleChoiceAction struct {
	Action
	Choices map[string]interface{} `json:"choices,omitempty"`
}

type TextEntryAction Action
type ConfirmAction Action
type DeleteAction Action
type CancelAction Action
