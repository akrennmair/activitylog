(
function() {

var PageVars = { };

$(document).ready(function() {

	var perform_login = function(result) {
		PageVars.authenticated = true;
		PageVars.activities = result.activities;
		$('#login_form').hide();
		$('#logged_in_form').show();
		//$('#logged_in_form #username').text(username);

		$('#login_page').hide();
		$('#main_page').show(200);

		$.get('/activity/latest', function(result) {
			var template = Handlebars.compile($('#tmpl_latest_activities_table').html());
			$('#latest_activities').html(template({activities: result}));
			$('.map-btn').click(function(e) {
				e.preventDefault();
				var coords = $(this).attr('data-coords').split(",");
				var latitude = parseFloat(coords[0]);
				var longitude = parseFloat(coords[1]);
				map.setView([latitude, longitude], 14);
				L.tileLayer('http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
					maxZoom: 18
				}).addTo(map);
				L.marker([latitude, longitude]).addTo(map);
				$('#modal_map').reveal();
			});
		});
	};

	var try_authenticate = function() {
		$.post('/auth/try', { }, function(result) {
			if (result.authenticated == true) {
				perform_login(result);
			}
		});
	};

	var map = L.map('map');

	$('#btn_signin').click(function(e) {
		e.preventDefault();
		var username = $('#login_form #username').val();
		var password = $('#login_form #password').val();
		$.post('/auth', { username: username, password: password }, function(result) {
			if (result.authenticated == true) {
				perform_login(result);
			} else {
				// TODO: show error message from result.errormsg
			}
		});
	});

	$('#btn_logout').click(function(e) {
		e.preventDefault();
		$.post('/auth/logout', { }, function(result) {
			$('#login_form').show();
			$('#logged_in_form').hide();

			$('#main_page').hide();
			$('#login_page').show(200);
		});
	});

	try_authenticate();

}); // $(document).ready


})();
