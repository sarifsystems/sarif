// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lua

const ModStore string = `
local store = {}

-- TODO: get rid when gopher-lua correctly supports coroutines
local function queue()
	local q, i, e

	local push = function(...)
		q = q or {}
		q[#q+1] = {...}
	end

	local pop = function()
		if not q then return end
		i, e = next(q, i)
		if not i then
			q = nil
			return
		end
		return unpack(e)
	end

	return push, pop
end

function store.scan(collection, filter)
	filter = filter or {}

	local push, pop = queue()
	return function()
		local k, v = pop()
		if k or v then
			return k, v
		end

		local msg = sarif.request{
			action = "store/scan/" .. collection,
			p = filter,
		}
		if not msg or not msg.p then return end

		for k, v in pairs(msg.p) do
			if filter.reverse then 
				if not filter['end'] or k < filter['end'] then
					filter['end'] = k:sub(0, -2)
				end
			else
				if not filter.start or k > filter.start then
					filter.start = k .. "~"
				end
			end
			push(k, v)
		end
		local a, b = pop()
		return a, b
	end
end

function store.get(key)
	local msg = sarif.request{action = "store/get/" .. key}
	return msg and msg.p
end

function store.put(key, val)
	local msg = sarif.request{action = "store/put/" .. key, p = val}
	return msg and msg.p and msg.p.key
end

return store
`
