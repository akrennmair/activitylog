(
function() {

var PageVars = { };

$(document).bind("pagebeforechange", function(event, data) {
	if (typeof data.toPage == 'object' && data.toPage.attr('data-needs-auth') == 'true') {
		if (!PageVars.authenticated) {
			event.preventDefault();
			$.mobile.changePage('#login_screen');
		}
	}
});

var log_activity = function(id, desc) {
	//console.log('sending activity ' + id);
	$.post('/activity/add', { "id": id, "desc": desc });
};

var get_latest_activities = function(id) {
	$.get('/activity/latest', null, function(activities) {
		$('#latest_activities_list li').remove();
		if (activities == null) {
			return;
		}
		for (var i=0;i<activities.length;i++) {
			var list_entry = $('<li></li>');
			list_entry.append(activities[i].ts + ': ' + activities[i].desc);
			$(id).append(list_entry);
		}
		$(id).listview("refresh").trigger('create');
	});
};

var populate_activities = function(id, activities) {
	//console.log("called populate_activities");
	for (var i=0;i<activities.length;i++) {
		var btn = $('<a data-role="button"></a>');
		btn.append(activities[i]);
		btn.click((function(id, desc) {
			return function() {
				log_activity(id, desc);
			}
		})(i, activities[i]));
		$(id).append(btn);
	}
	$(id).listview("refresh").trigger('create');
};

$(document).bind('pageinit', function() {
	$('#signin_btn').click(function() {
		var username = $('#username').val();
		var password = $('#password').val();

		$.post('/auth', { username: username, password: password }, function(result) {
			if (result.authenticated == true) {
				PageVars.authenticated = true;
				$.mobile.changePage('#page_submit');
				populate_activities('#submit_activity_list', result.activities);
			} else {
				$('#popup_error #msg').text(result.errormsg);
				$.mobile.changePage('#popup_error', { transition: "pop", role: "dialog" });
			}
		});

	});

	$('#show_latest').click(function() {
		get_latest_activities('#latest_activities_list');
	});

});

})();
