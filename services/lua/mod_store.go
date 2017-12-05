// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lua

const ModStore string = `
local store = {}

function store.scan(collection, filter)
	filter = filter or {}

	return coroutine.wrap(function()
		while true do
			local msg = sarif.request{
				action = "store/scan/" .. collection,
				p = filter,
			}
			if not msg or not msg.p or #msg.p.values == 0 then
				return
			end

			for i, v in ipairs(msg.p.values) do
				local k = msg.p.keys[i]
				if filter.reverse then 
					if not filter['end'] or k < filter['end'] then
						filter['end'] = k:sub(0, -2)
					end
				else
					if not filter.start or k > filter.start then
						filter.start = k .. "~"
					end
				end
				coroutine.yield(k, v)
			end
		end
	end)
end

function store.get(key)
	local msg = sarif.request{action = "store/get/" .. key}
	return msg and msg.p
end

function store.put(key, val)
	local msg = sarif.request{action = "store/put/" .. key, p = val}
	return msg and msg.p and msg.p.key
end

function store.batch(cmds)
	local msg = sarif.request{action = "store/batch", p = cmds}
	return msg and msg.p
end

return store
`
