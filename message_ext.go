package stark

func ReplyError(msg *Message, err error) *Message {
	reply := NewReply(msg)
	reply.Action = "error"
	if err != nil {
		reply.Message = err.Error()
		reply.Data["error"] = err.Error()
	}
	return reply
}

func ReplyUnsupported(msg *Message) *Message {
	reply := NewReply(msg)
	reply.Action = "error.unsupported"
	reply.Data["action"] = msg.Action
	reply.Message = "Cannot handle '" + msg.Action + "'"
	return reply
}
