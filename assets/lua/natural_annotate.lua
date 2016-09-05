sarif.subscribe("natural/annotate", "", function (msg)
	local p = msg.p

	if p and p.latitude and p.longitude then
		local latlng = p.latitude .. "," .. p.longitude
		p.attachments = p.attachments or {}
		p.attachments[#p.attachments+1] = {
			title = "View Location in Google Maps",
			title_link = "https://www.google.com/maps/place/" .. latlng,
			image_url = "https://maps.googleapis.com/maps/api/staticmap?zoom=17&size=600x300&markers=" .. latlng,
		}
	end

	sarif.reply(msg, {
		action = "natural/annotated",
		p = p,
	})
end)
