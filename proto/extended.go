package proto

func Subscribe(action, device string) Message {
	return Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": action,
			"device": device,
		},
	}
}
