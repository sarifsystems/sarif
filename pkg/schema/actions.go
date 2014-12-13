package schema

type Action struct {
	*Thing
	Reply interface{} `json:"reply,omitempty"`
	Name  string      `json:"name,omitempty"`
}

type TextEntryAction struct {
	Action
}

type MultipleChoiceAction struct {
	Action
	Choices map[string]interface{} `json:"choices,omitempty"`
}

type ConfirmAction Action
type DeleteAction Action
type CancelAction Action
