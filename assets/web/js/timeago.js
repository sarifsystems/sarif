angular.module('timeAgo', []).
	filter('timeago', function() {
		return function(input) {
			var substitute = function (stringOrFunction, number, strings) {
				var string = angular.isFunction(stringOrFunction) ? stringOrFunction(number, dateDifference) : stringOrFunction;
				var value = (strings.numbers && strings.numbers[number]) || number;
				return string.replace(/%d/i, value);
			},

			nowTime = (new Date()).getTime(),
			date = (new Date(input)).getTime(),
			strings = {
				prefixAgo: null,
				prefixFromNow: null,
				suffixAgo: "ago",
				suffixFromNow: "from now",
				seconds: "less than a minute",
				minute: "a minute",
				minutes: "%d minutes",
				hour: "an hour",
				hours: "%d hours",
				day: "a day",
				days: "%d days",
				month: "a month",
				months: "%d months",
				year: "a year",
				years: "%d years"
			},
			dateDifference = nowTime - date,
				words,
				seconds = Math.abs(dateDifference) / 1000,
				minutes = seconds / 60,
				hours = minutes / 60,
				days = hours / 24,
				years = days / 365,
				separator = strings.wordSeparator === undefined ?  " " : strings.wordSeparator,

			prefix = strings.prefixAgo,
			suffix = strings.suffixAgo;

			words = seconds < 45 && substitute(strings.seconds, Math.round(seconds), strings) ||
				seconds < 90 && substitute(strings.minute, 1, strings) ||
				minutes < 45 && substitute(strings.minutes, Math.round(minutes), strings) ||
				minutes < 90 && substitute(strings.hour, 1, strings) ||
				hours < 48 && substitute(strings.hours, Math.round(hours), strings) ||
				hours < 42 && substitute(strings.day, 1, strings) ||
				days < 30 && substitute(strings.days, Math.round(days), strings) ||
				days < 45 && substitute(strings.month, 1, strings) ||
				days < 365 && substitute(strings.months, Math.round(days / 30), strings) ||
				years < 1.5 && substitute(strings.year, 1, strings) ||
				substitute(strings.years, Math.round(years), strings);

			return [prefix, words, suffix].join(separator).trim();
		}
	});
