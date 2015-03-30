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

local scoring = {
	neutral = 0,

	positive = 2,
	negative = -1,
	angry = -2,
	sad = -2,

	relaxed = 1,
	anxious = -1,

	motivated = 1,
	unmotivated = -1,
}

local scales = {
	happiness = {"positive", "relaxed", "negative", "angry", "sad"},
	relaxation = {"relaxed", "anxious"},
	motivation = {"motivated", "unmotivated"},
}

local tags_by_mood = {
	neutral = {
		"neutral",
		"normal",
		"ok",
		"okay",
	},
	positive = {
		"happy",
		"good",
		"well",
		"proud",
		"optimistic",
		"awesome",
		"thrilled",
		"great",
		"liberated",
		"confident",
		"excited",
		"free",
	},
	relaxed = {
		"calm",
		"satisfied",
		"relieved",
		"relaxed",
		"comfortable",
		"content",
		"peaceful",
		"accomplished",
	},
	motivated = {
		"motivated",
		"energetic",
		"engaged",
		"occupied",
		"flow",
	},
	neutral = {
		"neutral",
		"normal",
		"okay",
		"ok",
	},
	negative = {
		"bad",
		"annoyed",
		"irritated",
		"disappointed",
		"discouraged",
		"ashamed",
		"powerless",
		"guilty",
		"sick",
		"grumpy",
		"disgruntled",
		"terrible",
		"embarrassed",
		"pessimistic",
		"jealous",
		"envious",
		"overwhelmed",
		"unsure",
		"uneasy",
		"humiliated",
		"desperate",
		"unhappy",
		"uncontent",
	},
	angry = {
		"angry",
		"hostile",
		"enraged",
		"upset",
		"hateful",
		"bitter",
		"infuriated",
	},
	sad = {
		"sad",
		"depressed",
		"hurt",
		"lost",
		"alone",
		"lonely",
		"vulnerable",
		"pathetic",
		"rejected",
		"heartbroken",
		"tearful",
		"sorrowful",
	},
	anxious = {
		"nervous",
		"anxious",
		"frightened",
		"shy",
		"tense",
		"paralyzed",
		"hesistant",
		"fearful",
		"terrified",
		"scared",
		"worried",
		"timid",
	},
	unmotivated = {
		"bored",
		"restless",
		"tired",
		"defeated",
		"frustrated",
	},
}

local tags = {}
for name, mood in pairs(tags_by_mood) do
	for _, tag in ipairs(mood) do
		tags[tag] = name
	end
end

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
	self:askSimple(msg, "How are you feeling?", "xp/mood")
end

function score_scale(scale, ...)
	local score = 0
	for _, mood in ipairs{...} do
		for _, accepted in pairs(scale) do
			if mood == accepted then
				score = score + scoring[mood]
			end
		end
	end
	return score
end

function analyze(text)
	if not text then return 0, "none", "none" end

	local score, primary, secondary = 0
	for word in string.gmatch(text, "([^, ]+)") do
		mood = tags[word]
		if mood then
			score = score + scoring[mood]
			if not primary then
				primary = mood
			elseif not secondary then
				secondary = mood
			end
		end
	end
	return score, primary or "neutral", secondary or "none"
end

function ExpSampling:recordMood(msg)
	local score, primary, secondary = analyze(msg.text)
	stark.publish{action = "mood/general/" .. primary .. "/" .. secondary .. "/" .. score, text = msg.text}

	for name, scale in pairs(scales) do
		stark.publish{action = "mood/scale/" .. name .. "/" .. score_scale(scale, primary, secondary)}
	end

	--self:askSimple(msg, "How happy do you feel?", "xp/happiness")
	self:askSimple(msg, "What are you doing?", "xp/activity")
end

function ExpSampling:analyze(msg)
	local score, primary, secondary = analyze(msg.text)
	stark.reply(msg, {action = "xp/analyzed",
		text = score .. " " .. primary .. " " .. secondary,
		p = {
			score = score,
			primary = primary,
			secondary = secondary,
		},
	})
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
	local facts = stark.request{action = "cmd/catfacts"}
	if facts then
		facts = "\nHere's a cat fact: " .. facts.text
	end
	stark.reply(msg, {action = "xp/done", text = "Thanks!" .. (facts or " ")})


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
stark.subscribe("xp/analyze",  "", function(msg) ExpSampling:analyze(msg) end)
stark.subscribe("cmd/mood",  "", function(msg) ExpSampling:analyze(msg) end)
stark.subscribe("xp/ask", "", function(msg) ExpSampling:startAsking(msg) end)
stark.subscribe("xp/mood", "", function(msg) ExpSampling:recordMood(msg) end)
stark.subscribe("xp/happiness", "", function(msg) ExpSampling:recordHappiness(msg) end)
stark.subscribe("xp/relaxation",  "", function(msg) ExpSampling:recordRelaxation(msg) end)
stark.subscribe("xp/motivation",  "", function(msg) ExpSampling:recordMotivation(msg) end)
stark.subscribe("xp/activity",  "", function(msg) ExpSampling:recordActivity(msg) end)

stark.subscribe("cron/05h",  "", function(msg) ExpSampling:scheduleToday() end)
