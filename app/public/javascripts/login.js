$(function() {

	if (Cookies.get('token') != undefined) {
		Cookies.remove('token');
		document.location.reload(true);
	} else if (getUrlVars()['token'] != undefined) {

		if (window.location.href[window.location.href.length - 1] === '#') {
			window.location.replace(window.location.href.substring(0, window.location.href.length - 1));
			setTimeout(function() { 
				Cookies.set('token', getUrlVars()['token']);
				window.location.href = "/";
			}, 350);
		} else {
			Cookies.set('token', getUrlVars()['token']);
			window.location.href = "/";
		}
	}

});

$('#inputLogInGoogle').click(function() {
	window.location.replace('http://localhost:9090/api/auth/google');
});

$('#inputLogIn42').click(function() {
	window.location.replace('http://localhost:9090/api/auth/fortytwo');
});

$('#loginForm').submit(function(e) {

	var username = $('#inputLoginUsername');
	var password = $('#inputLoginPassword');

	removeInputDangerClass(username);
	removeInputDangerClass(password);

	if (username.val().length < 8 || username.val().length > 15 || specialCharRegex.test(username.val()) === false) {
		addInputDangerClass(username);
		e.preventDefault();
	}
	if (password.val().length <= 8 || password.val().length >= 64 || passwordRegex.test(password.val())=== false) {
		addInputDangerClass(password);
		e.preventDefault();
	}

	var formData = new FormData($(this)[0]);

	$.ajax({
		url: 'http://localhost:9090/api/auth/login',
		type: 'POST',
		data: formData,
		success: function (data) {
			addFlashMessage('#loginForm', 'Congratulations', 'You have successfully loged in', 'success');
			window.location.replace("/login?token=" + data["token"]);
		},
		error: function (data) {
			addFlashMessage('#loginForm', data.statusText, data.responseText, 'danger');
		},
		cache: false,
		contentType: false,
		processData: false
	});
});