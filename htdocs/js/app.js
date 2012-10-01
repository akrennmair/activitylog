(function($) {

	var PageVars = { };

	var map = L.map('map');

	var app = $.sammy(function() {

		var show_page = function(id) {
			$('.container-fluid').hide();
			$('#navbar').show();
			$(id).show();
			console.log('showing ' + id);
		};

		var show_subpage = function(id) {
			$('.subpage').hide();
			$(id).show();
			console.log('showing subpage ' + id);
		};

		var update_activities_map = function() {
			PageVars.activities_map = { };
			for (var i=0;i<PageVars.activities.length;i++) {
				PageVars.activities_map[PageVars.activities[i].type_id] = PageVars.activities[i];
			}
		};

		var perform_login = function(result) {
			PageVars.authenticated = true;
			PageVars.activities = result.activities;
			update_activities_map();
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
		};

		var bind_edit_and_delete_buttons = function() {
			$('#edit_activity_types .edit_btn').click(function() {
				var id = $(this).attr('data-id');
				var old_edit_row = $('#at_edit_' + id).clone();

				var template = Handlebars.compile($('#tmpl_edit_activity_type').html());
				$('#at_edit_' + id).replaceWith(template({type_id: id, name: PageVars.activities_map[id].name}));

				$('#at_edit_' + id + ' #btn_cancel').click(function() {
					$('#at_edit_' + id).replaceWith(old_edit_row);
					bind_edit_and_delete_buttons();
				});

				$('#at_edit_' + id + ' #btn_save').click(function() {
					var new_name = $('#at_edit_' + id + ' #new_name').val();
					$.post('/activity/type/edit', { id: id, newname: new_name }, function(result) {
						PageVars.activities_map[id].name = new_name;
						var template = Handlebars.compile($('#tmpl_activity_type_row').html());
						$('#at_edit_' + id).replaceWith(template({type_id: id, name: new_name}));
						bind_edit_and_delete_buttons();
					});
				});

			});
			$('#edit_activity_types .delete_btn').click(function() {
				var id = $(this).attr('data-id');
				console.log('delete activity type ' + id);
				if (window.confirm("Are you sure you want to delete this activity type?") === true) {
					$.post('/activity/type/del', { id: id }, function(result) {
						$('#at_edit_' + id).remove();
					});
				}
			});
		};

		var load_activity_types = function() {
			$.get('/activity/type/list', function(result) {
				PageVars.activities = result;
				update_activities_map();
				var template = Handlebars.compile($('#tmpl_edit_activity_types').html());
				$('#edit_activity_types').html(template({activity_types: result}));
				bind_edit_and_delete_buttons();
			});
		};

		this.get('#/activity/types/edit', function(ctx) {
			show_subpage('#edit_activity_types');

			load_activity_types();
		});

		this.get('#/about', function(ctx) {
			show_page('#about_page');
		});

		this.get('#/contact', function(ctx) {
			show_page('#contact_page');
		});

		this.get('#/main', function(ctx) {
			if (!PageVars.authenticated) {
				ctx.redirect("#/");
				return;
			}

			show_page('#main_page');
			show_subpage('#latest_activities');
			load_latest_activities();
		});

		this.get('#/logout', function(ctx) {
			$.post('/auth/logout', { dummy: "ignore" }, function(result) {
				$('#login_form').show();
				$('#logged_in_form').hide();

				ctx.redirect('#/');
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
			show_page('#login_page');
			try_authenticate();
		});
	});

	$(function() {
		app.run('#/');
	});

})(jQuery);
