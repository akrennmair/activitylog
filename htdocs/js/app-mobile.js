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
	PageVars.type_id = id;
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

var add_activity_btn = function(list_id, id, desc) {
	var btn = $('<a data-role="button" data-icon="arrow-r" data-iconpos="right"></a>');
	btn.text(desc);
	btn.click(function() {
		goto_submit_page(id, desc);
	});
	$(list_id).append(btn);
};

var populate_activities = function(id, activities) {
	for (var i=0;i<activities.length;i++) {
		add_activity_btn(id, activities[i].type_id, activities[i].name);
	}
	$(id).listview("refresh").trigger('create');
};

$(document).ready(function() {
	$('#signin_btn').click(function(e) {
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

	var add_activity = function(data, f) {
		$.post('/activity/add', data, function() {
			$.mobile.changePage('#page_submit');
		});
	};

	$('#submit_activity_btn').click(function() {
		var type_id = PageVars.type_id;
		var desc = $('#submit_detail_view #description').val();
		var is_public = $('#submit_detail_view #public').val();
		var recordloc = $('#submit_detail_view #recordloc').val();
		if (recordloc == "on") {
			if ("geolocation" in navigator) {
				navigator.geolocation.getCurrentPosition(function(position) {
					add_activity({ "type_id": type_id, "desc": desc, "public": is_public, 'lat': position.coords.latitude, 'long': position.coords.longitude });
				});
				return;
			}
		}
		add_activity({ "type_id": type_id, "desc": desc, "public": is_public });
	});

	$('#add_activity_type_btn').click(function() {
		var typename = $('#activity_add_view #typename').val();
		$.post('/activity/type/add', { "typename": typename }, function(result) {
			add_activity_btn('#submit_activity_list', result.type_id, typename);
			$('#submit_activity_list').listview("refresh").trigger('create');
			$('#activity_add_view #typename').val("");
			$.mobile.changePage('#page_submit');
		});
	});
});

})();
