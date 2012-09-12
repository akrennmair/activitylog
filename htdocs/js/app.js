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
	});
};

var populate_activities = function(id, activities) {
	for (var i=0;i<activities.length;i++) {
		var btn = $('<a class="btn btn-block"></a>');
		btn.append(activities[i]);
		btn.click((function(id) {
			return function() {
				log_activity(id);
			}
		})(i));
		$(id).append(btn);
	}
};

$(document).ready(function() {
	populate_activities('#submit_activity_list', activities);

	$('#show_latest').click(function() {
		$('#submit_view').hide();
		$('#log_view').show();
		get_latest_activities('#latest_activities_list');
	});

	$('#submit_activity').click(function() {
		$('#log_view').hide();
		$('#submit_view').show();
	});
});

})();
