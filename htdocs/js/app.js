(
function() {

var PageVars = { };

$(document).ready(function() {

	$('#btn_signin').click(function(e) {
		e.preventDefault();
		var username = $('#login_form #username').val();
		var password = $('#login_form #password').val();
		$.post('/auth', { username: username, password: password }, function(result) {
			if (result.authenticated == true) {
				PageVars.authenticated = true;
				$('#login_form').hide();
				$('#logged_in_form').show();
				//$('#logged_in_form #username').text(username);

				$('#login_page').hide();
				$('#main_page').show(200);
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

});


})();
