function StarkClient(deviceId) {
	this.deviceId = deviceId;
}

StarkClient.prototype.Connect = function(msg) {
	this.socket = new WebSocket("ws://" + window.location.host + "/stream/stark");
	var client = this

	this.socket.onopen = function() {
		console.log('open');
		if (client.onOpen) {
			client.onOpen();
		}
	}

	this.socket.onmessage = function(raw) {
		console.log("receive", raw.data);
		msg = JSON.parse(raw.data);
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
	msg.v = msg.v || "0.3"
	msg.id = msg.id || GenerateId()
	msg.src = msg.src || this.deviceId

	raw = JSON.stringify(msg);
	console.log("publish", raw)
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

function GenerateId() {
    var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

    var text = "";
    for( var i = 0; i < 8; i++ )
        text += possible.charAt(Math.floor(Math.random() * possible.length));

    return text;
}
