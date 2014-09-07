var mainApp = angular.module('mainApp', []);

mainApp.controller('ChatCtrl', function ($scope) {
	$scope.responses = [];

	s = new StarkClient("web-" + GenerateId());
	s.onMessage = function(msg) {
		if (msg.action == "ping") {
			s.Publish({
				action: "ack",
				dst: msg.src,
				corr: msg.id,
			});
		}

		$scope.$apply(function() {
			$scope.addMessage(msg);
		});
	};
	s.onOpen = function(msg) {
		s.Subscribe("ping", "");
		s.Subscribe("", "self");
	};
	s.Connect()

	$scope.publish = function() {
		if (!$scope.message) {
			return;
		}
		if ($scope.message == "clear") {
			$scope.responses = [];
		} else {
			var msg = {
				action: "natural/handle",
				p: {
					text: $scope.message,
				}
			};
			s.Publish(msg);
			$scope.addMessage(msg);
		}
		$scope.message = "";
	};

	$scope.addMessage = function(msg) {
		var chat = {
			msg: msg,
			isSelf: (msg.src == s.deviceId),
			time: new Date(),
			text: msg.p && msg.p.text || (msg.action + " from " + msg.src)
		}
		$scope.responses.push(chat);
	};
});
