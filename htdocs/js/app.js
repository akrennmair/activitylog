(
function() {

var activities = [ 'Eat', 'Sleep', 'Drink', 'Shopping' ];

var log_activity = function(id) {
	console.log('sending activity ' + id);
	$.post('/activity/add', { "id": id, "desc": activities[id] });
};

var get_latest_activities = function(id) {
	$.get('/activity/latest', null, function(activities) {
		$('#latest_activities_list li').remove();
		for (var i=0;i<activities.length;i++) {
			var list_entry = $('<li></li>');
			list_entry.append(activities[i].ts + ': ' + activities[i].desc);
			$(id).append(list_entry);
		}
		$(id).listview("refresh").trigger('create');
	});
};

var populate_activities = function(id, activities) {
	for (var i=0;i<activities.length;i++) {
		var btn = $('<a data-role="button"></a>');
		btn.append(activities[i]);
		btn.click((function(id) {
			return function() {
				log_activity(id);
			}
		})(i));
		$(id).append(btn);
	}
	$(id).listview("refresh").trigger('create');
};

$(document).bind('pageinit', function() {
	$('#signin_btn').click(function() {
		$.mobile.changePage('#page_submit');
		populate_activities('#submit_activity_list', activities);
	});

	$('#show_latest').click(function() {
		get_latest_activities('#latest_activities_list');
	});

});

})();
