local stark = require "stark"

local ExpSampling = {}

ExpSampling.Times = {
	{9, 10},

	{11, 13},
	{14, 16},
	{17, 19},
	{20, 22},

	{23, 0},
}

function ExpSampling:askSimple(msg, question, reply_action)
	local r = {
		action = "xp/question",
		text = question,
		p = {
			action = {
				["@type"] = "TextEntryAction",
				reply = reply_action,
			},
		},
	}
	if msg then
		stark.reply(msg, r)
	else
		r.dest = "user"
		stark.publish(r)
	end
end

function ExpSampling:startAsking(msg)
	local facts = stark.request{action = "cmd/catfacts"}
	if facts then
		facts = "\nHere's a cat fact: " .. facts.text
	end
	stark.reply(msg, {action = "xp/done", text = "Thanks!" .. (facts or " ")})

	--self:askSimple(msg, "How happy do you feel?", "xp/happiness")
end

function ExpSampling:recordHappiness(msg)
	self:askSimple(msg, "How relaxed do you feel?", "xp/relaxation")

	stark.publish{action = "mood/happiness/" .. msg.text}
end

function ExpSampling:recordRelaxation(msg)
	self:askSimple(msg, "How motivated are you?", "xp/motivation")

	stark.publish{action = "mood/relaxation/" .. msg.text}
end

function ExpSampling:recordMotivation(msg)
	self:askSimple(msg, "What are you doing?", "xp/activity")

	stark.publish{action = "mood/motivation/" .. msg.text}
end

function ExpSampling:recordActivity(msg)
	stark.reply(msg, {action = "xp/done", text = "Thanks!"})

	stark.publish{action = "activity", text = msg.text}
end

function ExpSampling:scheduleToday()
	local currhour = os.date("*t").hour
	for _, t in ipairs(self.Times) do
		if t[1] < 5 or t[1] > currhour then
			stark.publish{
				action = "schedule",
				p = {
					reply = {action = "xp/start"},
					random_after = t[1] .. ":00",
					random_before = t[2] .. ":00",
				},
			}
		end
	end
end

stark.subscribe("xp/start",  "", function(msg) ExpSampling:startAsking() end)
stark.subscribe("xp/ask", "", function(msg) ExpSampling:startAsking(msg) end)
stark.subscribe("xp/happiness", "", function(msg) ExpSampling:recordHappiness(msg) end)
stark.subscribe("xp/relaxation",  "", function(msg) ExpSampling:recordRelaxation(msg) end)
stark.subscribe("xp/motivation",  "", function(msg) ExpSampling:recordMotivation(msg) end)
stark.subscribe("xp/activity",  "", function(msg) ExpSampling:recordActivity(msg) end)

stark.subscribe("cron/5h",  "", function(msg) ExpSampling:scheduleToday() end)
