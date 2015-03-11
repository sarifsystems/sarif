local stark = require "stark"

stark.subscribe("lua/examples/echo", "", function(msg)
	stark.reply(msg, {
		action = "echoed",
		text = msg.text,
	})
end)

stark.subscribe("cmd/calc", "", function(msg)
	if not msg.text or msg.text == "" then
		return stark.reply_error(msg, "badrequest", "no expression")
	end

	local v = loadstring("return " .. msg.text)
	local ok, ret = pcall(v)
	if not ok then return stark.reply_error(msg, "badrequest", ret) end

	return stark.reply(msg, {
		action = "calculated",
		text = ret,
	})
end)
