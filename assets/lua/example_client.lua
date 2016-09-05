local sarif = require "sarif"

sarif.subscribe("lua/examples/echo", "", function(msg)
	sarif.reply(msg, {
		action = "echoed",
		text = msg.text,
	})
end)

sarif.subscribe("cmd/calc", "", function(msg)
	if not msg.text or msg.text == "" then
		return sarif.reply_error(msg, "badrequest", "no expression")
	end

	local v = loadstring("return " .. msg.text)
	local ok, ret = pcall(v)
	if not ok then return sarif.reply_error(msg, "badrequest", ret) end

	return sarif.reply(msg, {
		action = "calculated",
		text = ret,
	})
end)
