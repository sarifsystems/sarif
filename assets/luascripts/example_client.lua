local function reply(msg, r)
	r.dest = msg.src
	r.corr = msg.id
	return publish(r)
end

local function replyerr(msg, t, err)
	return reply{
		action = "err/" .. t,
		text = err,
	}
end

subscribe("lua/examples/echo", "", function(msg)
	reply(msg, {
		action = "echoed",
		text = msg.text,
	})
end)

subscribe("cmd/calc", "", function(msg)
	if not msg.text or msg.text == "" then
		return replyerr(msg, "badrequest", "no expression")
	end

	local v = loadstring("return " .. msg.text)
	local ok, ret = pcall(v)
	if not ok then return replyerr(msg, "badrequest", ret) end

	return reply(msg, {
		action = "calculated",
		text = ret,
	})
end)
