local sarif = require "sarif"

local Analyzer = {}

local hex_to_char = function(x)
  return string.char(tonumber(x, 16))
end

function dump(o)
   if type(o) == 'table' then
      local s = '{ '
      for k,v in pairs(o) do
         if type(k) ~= 'number' then k = '"'..k..'"' end
         s = s .. '['..k..'] = ' .. dump(v) .. ','
      end
      return s .. '} '
   else
      return tostring(o)
   end
end

function Analyzer:reset()
	if self.numGoogles and self.numGoogles > 0 then
		sarif.publish{
			action = "report/daily/google",
			p = {
				count = self.numGoogles,
				queries = self.queries,
			},
		}
	end
	self.numGoogles = 0
	self.queries = {}
end

function Analyzer:handleUpdate(msg)
	for url, title in pairs(msg.p.visited_urls) do
		local query = url:match("google%.%a+/search?.*q=([^&]+)")
		if query then
			self.numGoogles = self.numGoogles + 1
			query = query:gsub("%%(%x%x)", hex_to_char)
			self.queries[#self.queries+1] = query

			sarif.publish{
				action = "browser/daily/google",
				p = {
					count = self.numGoogles,
					query = query,
				},
			}
		end
	end
end

function Analyzer:handleStats(msg)
	sarif.reply(msg, {
		action = "browser/daily/got_stats",
		p = {
			count = self.numGoogles,
			queries = self.queries,
		},
	})
end

sarif.subscribe("browser/session/update",  "", function(msg) Analyzer:handleUpdate(msg) end)
sarif.subscribe("browser/daily/stats",  "", function(msg) Analyzer:handleStats(msg) end)

sarif.subscribe("cron/05h",  "", function(msg) Analyzer:reset() end)

Analyzer:reset()
