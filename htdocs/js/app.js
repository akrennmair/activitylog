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

var goto_submit_page = function(id, desc) {
	PageVars.activity_id = id;
	$('#submit_detail_view #description').val(desc);
	$.mobile.changePage('#page_submit_detail');
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
				goto_submit_page(id, desc);
			}
		})(i, activities[i]));
		$(id).append(btn);
	}
	$(id).listview("refresh").trigger('create');
};

$(document).ready(function() {
	$('#signin_btn').click(function() {
		var username = $('#login_screen #username').val();
		var password = $('#login_screen #password').val();

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

	$('#signup_btn').click(function() {
		var username = $('#signup_screen #username').val();
		var password1 = $('#signup_screen #password1').val();
		var password2 = $('#signup_screen #password2').val();

		if (password1 != password2) {
			$('#popup_error #msg').text("Passwords do not match.");
			$.mobile.changePage('#popup_error', { transition: "pop", role: "dialog" });
			return;
		}

		if (password1.length == 0) {
			$('#popup_error #msg').text("Password is empty.");
			$.mobile.changePage('#popup_error', { transition: "pop", role: "dialog" });
			return;
		}

		$.post('/auth/signup', { "username": username, "password": password1 }, function(result) {
			if (result.authenticated == true) {
				$.mobile.changePage('#login_screen');
			} else {
				$('#popup_error #msg').text(result.errormsg);
				$.mobile.changePage('#popup_error', { transition: "pop", role: "dialog" });
			}
		});
	});

	$('#submit_activity_btn').click(function() {
		var id = PageVars.activity_id;
		var desc = $('#submit_detail_view #description').val();
		$.post('/activity/add', { "id": id, "desc": desc }, function() {
			$.mobile.changePage('#page_submit');
		});
	});
});

})();
