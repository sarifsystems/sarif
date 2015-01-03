var mainApp = angular.module('mainApp', ['timeAgo', 'ui.bootstrap']);

mainApp.factory('stark', function($rootScope, $q) {
	var client = new StarkClient("web-" + GenerateId());
	client.Request = function(msg, onReply) {
		if (onReply) {
			StarkClient.prototype.Request.call(client, msg, function(msg) {
				onReply(msg);
				$rootScope.$digest();
			});
		} else {
			var deferred = $q.defer();
			StarkClient.prototype.Request.call(client, msg, function(msg) {
				deferred.resolve(msg);
				$rootScope.$digest();
			});
			return deferred.promise;
		}
	};
	return client;
});

mainApp.controller('ChatCtrl', function ($scope, stark) {
	$scope.responses = [];
	$scope.dropdownOpen = false;

	$scope.publish = function() {
		if (!$scope.message) {
			return;
		}
		if ($scope.message == "clear") {
			$scope.responses = [];
		} else {
			var msg = {
				action: "natural/handle",
				text: $scope.message,
			};
			stark.Request(msg, $scope.addMessage);
			$scope.addMessage(msg);
		}
		$scope.message = "";
	};

	$scope.openDropdown = function() {
		$scope.dropdownOpen = true;
	}

	$scope.addMessage = function(msg) {
		var chat = {
			msg: msg,
			isSelf: (msg.src == stark.deviceId),
			time: new Date(),
			text: msg.text || (msg.action + " from " + msg.src)
		}
		$scope.responses.push(chat);
		$scope.openDropdown();
	};
});

mainApp.controller('DashboardCtrl', function($scope, stark) {
	stark.Request({
		action: 'event/last',
		p: {
			verb: 'drink',
			object: 'coffee'
		}
	}).then(function(msg) {
		$scope.lastCoffee = msg.p;
	});

	stark.Request({
		action: 'location/last',
	}).then(function(msg) {
		$scope.lastLocation = msg.p;
	});
});
