package stark

// ReplyError generates a new error reply to the passed Message.
// 
func ReplyError(msg *Message, err error) *Message {
	reply := NewReply(msg)
	reply.Action = "error"
	if err != nil {
		reply.Message = err.Error()
		reply.Data["error"] = err.Error()
	}
	return reply
}

// ReplyUnsupported generates a new error.unsupported reply to the passed Message.
// This is often used to signal that the replying service cannot handle the
// specified action.
func ReplyUnsupported(msg *Message) *Message {
	reply := NewReply(msg)
	reply.Action = "error.unsupported"
	reply.Data["action"] = msg.Action
	reply.Message = "Cannot handle '" + msg.Action + "'"
	return reply
}
