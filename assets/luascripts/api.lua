local sarif = require "sarif"

sarif.subscribe("btc/get_price", "", function(msg)
	price = sarif.request{
		action = "json/api.bitfinex.com/v1/pubticker/btcusd",
		p = {},
	}
	local p = price.p.result
	p.value = p.last_price
	local reply = {
		action = "btc/price",
		text = "The last bitcoin price is " .. p.last_price .. ".",
		p = p,
	}
	sarif.reply(msg, reply)
	sarif.publish(reply)
end)
