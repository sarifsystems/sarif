local stark = require "stark"

local function formatTime(t)
	local d = os.date("!%Y-%m-%dT%H:%M:%S%z", t or os.time())
	return d:sub(1, -3) .. ":" .. d:sub(-2)
end

local Tracker = {}

function Tracker:start(msg)
	local category = msg.p and msg.p.category or "unspecified"
	local text = msg.p and msg.p.text or msg.text

	local event = stark.request{
		action = "event/new",
		text = category .. ": " .. text,
		p = {
			action = "timetracker/activity/" .. category .. "/start",
		},
	}
	stark.reply(msg, {
		action = "timetracker/tracked",
		p = event.p,
		text = event.text,
	})
end

function Tracker:fetchActive()
	local last = stark.request{
		action = "event/last",
		p = {
			action = "timetracker/activity",
		},
	}
	if not last or not last.p or not last.p.action then
		return nil
	end
	local category = string.match(last.p.action, "^timetracker/activity/(.+)/start$")
	return category and last, category
end

function Tracker:active(msg)
	local last, category = self:fetchActive()
	if not category then
		return stark.reply_error(msg, "notfound", "No active time found")
	end
	stark.reply(msg, {
		action = "timetracker/active",
		text = "Currently tracking " .. last.text,
	})
end

function Tracker:stop(msg)
	local last, category = self:fetchActive()
	if not category then
		return stark.reply_error(msg, "notfound", "No active time found")
	end
	local text = "Stopped tracking " .. last.text,

	stark.request{
		action = "event/new",
		text = text,
		p = {
			action = "timetracker/activity/" .. category .. "/stop",
		},
	}
	stark.reply(msg, {
		action = "timetracker/stopped",
		text = text,
	})
end

function Tracker:today(msg)
	local category = msg.p and msg.p.category or ""
	local events = stark.request{
		action = "event/list",
		p = {
			action = "timetracker/activity/" .. category,
			after = formatTime(os.time() - 86400),
		}
	}
	stark.reply(msg, {
		action = "timetracker/today",
		text = events.text,
		p = events.p,
	})
end

stark.subscribe("timetracker/start", "", function(msg) Tracker:start(msg) end)
stark.subscribe("timetracker/stop", "", function(msg) Tracker:stop(msg) end)
stark.subscribe("timetracker/active", "", function(msg) Tracker:active(msg) end)
stark.subscribe("timetracker/today", "", function(msg) Tracker:today(msg) end)
