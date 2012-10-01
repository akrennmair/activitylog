(function($) {

	var PageVars = { };

	var map = L.map('map');

	var app = $.sammy(function() {

		var hide_all_pages = function() {
			$('.container-fluid').hide();
			$('#navbar').show();
		};

		var perform_login = function(result) {
			PageVars.authenticated = true;
			PageVars.activities = result.activities;
			$('#login_form').hide();
			$('#logged_in_form').show();
			//$('#logged_in_form #username').text(username);
			
			app.setLocation('#/main');
		};

		var try_authenticate = function() {
			$.post('/auth/try', { dummy: "ignored" }, function(result) {
				if (result.authenticated == true) {
					perform_login(result);
				}
			});
		};

		var load_latest_activities = function() {
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
		}

		this.get('#/about', function(ctx) {
			hide_all_pages();
			$('#about_page').show();
		});

		this.get('#/contact', function(ctx) {
			hide_all_pages();
			$('#contact_page').show();
		});

		this.get('#/main', function(ctx) {
			if (!PageVars.authenticated) {
				app.setLocation("#/");
				return;
			}

			hide_all_pages();
			$('#main_page').show();
			load_latest_activities();
		});

		this.get('#/logout', function(ctx) {
			$.post('/auth/logout', { dummy: "ignore" }, function(result) {
				$('#login_form').show();
				$('#logged_in_form').hide();

				app.setLocation('#/');
			});
		});

		this.post('#/auth', function(ctx) {
			var username = $('#login_form #username').val();
			var password = $('#login_form #password').val();
			$.post('/auth', { username: username, password: password }, function(result) {
				if (result.authenticated == true) {
					perform_login(result);
				} else {
					var template = Handlebars.compile($('#tmpl_errormsg').html());
					$('#errormsg_placeholder').html(template({message: result.errormsg}));
				}
			});
			return false;
		});

		this.get('#/', function(ctx) {
			$('#errormsg_placeholder').html("");
			hide_all_pages();
			$('#login_page').show();
			try_authenticate();
		});
	});

	$(function() {
		app.run('#/');
	});

})(jQuery);
