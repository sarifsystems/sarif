local stark = require "stark"

stark.subscribe("btc/get_price", "", function(msg)
	price = stark.request{
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
	stark.reply(msg, reply)
	stark.publish(reply)
end)
