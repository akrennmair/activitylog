(function($) {

	var PageVars = { };

	var map = L.map('map');

	var load_next_page = function() {
		if (!PageVars.latest_activities_reached_end) {
			$.get('/activity/list/' + PageVars.latest_activities_page, function(data) {
				if (data.length == 0) {
					console.log('reached end page = ' + PageVars.latest_activities_page);
					PageVars.latest_activities_reached_end = true;
				} else {
					console.log('loaded page ' + PageVars.latest_activities_page);
					PageVars.latest_activities_page++;
					var template = Handlebars.compile($('#tmpl_latest_activities_table').html());
					var new_rows = template({activities: data});
					$('#latest_activities_table').append(new_rows);
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
				}
			});
		}
	};

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
			PageVars.latest_activities_reached_end = false;
			PageVars.latest_activities_page = 1;
			load_next_page();
		};

		var bind_edit_and_delete_buttons = function() {
			$('#edit_activity_types .edit_btn').unbind('click');
			$('#edit_activity_types .edit_btn').click(function() {
				var id = $(this).attr('data-id');
				var old_edit_row = $('#at_edit_' + id).clone();

				var template = Handlebars.compile($('#tmpl_edit_activity_type').html());
				$('#at_edit_' + id).replaceWith(template({type_id: id, name: PageVars.activities_map[id].name}));

				$('#at_edit_' + id + ' #btn_cancel').click(function() {
					var edit_row = old_edit_row;
					$('#at_edit_' + id).replaceWith(edit_row);
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
			$('#edit_activity_types .deletdeletee_btn').unbind('click');
			$('#edit_activity_types .delete_btn').click(function() {
				var id = $(this).attr('data-id');
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
				$('#add_activity_type_btn').click(function(e) {
					e.preventDefault();
					var name = $('#at_newname').val();
					var is_time_period = $('#is_time_period').is(':checked');
					console.log('is_time_period = ' + is_time_period);
					$.post('/activity/type/add', { "typename": name, "time_period": is_time_period }, function(result) {
						PageVars.activities.push(result);
						update_activities_map();
						var tmpl = Handlebars.compile($('#tmpl_activity_type_row').html());
						$('#at_newname').val("");
						$('#edit_activity_types_tbody').append(tmpl(result));
						bind_edit_and_delete_buttons();
					});
				});
				bind_edit_and_delete_buttons();
			});
		};

		var add_activity = function(data, f) {
			$.post('/activity/add', data, f);
		};

		var show_add_activity_page = function(ctx) {
			var template = Handlebars.compile($('#tmpl_add_activity').html());
			$('#add_activity').html(template({activity_types: PageVars.activities}));
			$('#add_activity #add_activity_btn').click(function(e) {
				e.preventDefault();
				var activity_type_id = $('#add_activity_form #activity_type').val();
				var description = $('#add_activity_form #description').val();
				var is_public = $('#add_activity_form #public').val();
				// TODO: also record current location
				add_activity({ "type_id": activity_type_id, "desc": description, "public": is_public }, function() {
					ctx.redirect('#/main');
				});
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

		this.get('#/activity/add', function(ctx) {
				show_page('#main_page');
				show_subpage('#add_activity');

				show_add_activity_page(ctx);
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

	$(document).ready(function() {
		PageVars.latest_activities_reached_end = false;
		PageVars.latest_activities_page = 1;
		$('#latest_activities').waypoint(function(event, direction) {
			console.log('waypoint event! dir = ' + direction);
			if (direction === 'down') {
				load_next_page();
			}
		}, { offset: 'bottom-in-view' });
	});

})(jQuery);
