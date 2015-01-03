function StarkClient(deviceId) {
	this.deviceId = deviceId;
	this.replyHandlers = {};
	this.pubQueue = [];

	this.socket = new WebSocket("ws://" + window.location.host + "/stream/stark");
	var client = this

	this.socket.onopen = function() {
		console.log('open');
		client.Subscribe("ping", "");
		client.Subscribe("", "self");

		while (raw = client.pubQueue.pop()) {
			console.log("publishq", raw);
			this.send(raw);
		}
		if (client.onOpen) {
			client.onOpen();
		}
	}

	this.socket.onmessage = function(raw) {
		console.log("receive", raw.data);
		var msg = JSON.parse(raw.data);

		if (msg.action == "ping") {
			client.Publish({
				action: "ack",
				dst: msg.src,
				corr: msg.id,
			});
		}
		if (msg.corr) {
			var handler = client.replyHandlers[msg.corr]
			if (handler) {
				return handler(msg)
			}
		}
		if (client.onMessage) {
			client.onMessage(msg);
		}
	}

	this.socket.onclose = function() {
		if (client.onClose) {
			client.onClose();
		}
		console.log('closed');
	}
}

StarkClient.prototype.Publish = function(msg) {
	msg.stark = msg.stark || "0.5"
	msg.id = msg.id || GenerateId()
	msg.src = msg.src || this.deviceId

	raw = JSON.stringify(msg);
	if (this.socket.readyState != WebSocket.OPEN) {
		this.pubQueue.push(raw);
		return;
	}
	console.log("publish", raw);
	this.socket.send(raw);
}

StarkClient.prototype.Subscribe = function(action, device) {
	if (!device) {
		this.Subscribe(action, this.deviceId)
	}
	var msg = {
		action: "proto/sub",
		p: {
			action: action
		}
	}
	if (device) {
		msg.p.device = (device == "self" ? this.deviceId : device)
	}
	this.Publish(msg)
}

StarkClient.prototype.Request = function(msg, onReply) {
	msg.id = msg.id || GenerateId()
	this.replyHandlers[msg.id] = onReply;
	var client = this
	window.setTimeout(function() {
		delete client.replyHandlers[msg.id];
	}, 300000)
	return this.Publish(msg)
}

function GenerateId() {
    var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

    var text = "";
    for( var i = 0; i < 8; i++ )
        text += possible.charAt(Math.floor(Math.random() * possible.length));

    return text;
}
