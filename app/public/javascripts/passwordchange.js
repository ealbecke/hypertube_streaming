$(function() {

	$('#inputNewPassword').focusout(function(e) {
		removeFeedbackMessage(this)
		removeInputDangerClass(this);
		removeInputSuccessClass(this);
		if ($(this).val().length > 0) {
			if ($(this).val().length < 8 || $(this).val().length > 64 || (passwordRegex.test($(this).val()) === false)) {
				addInputDangerClass(this);
				addFeedbackMessage(this, messageErrorPasswordSize + ', '+ messageErrorPasswordRegExp);
			} else {
				addInputSuccessClass(this);
			}
		}
	});
	$('#inputConfirmNewPassword').focusout(function(e) {
		removeFeedbackMessage(this)
		removeInputDangerClass(this);
		removeInputSuccessClass(this);
		if ($(this).val().length > 0) {
			if ($(this).val() != $('#inputNewPassword').val()) {
				addInputDangerClass(this);
				addFeedbackMessage(this, messageErrorConfirmPassword);
			} else {
				addInputSuccessClass(this);
			}
		}
	});

});

$('#passwordChange').submit(function(e) {

	e.preventDefault();

	var password = $('#inputNewPassword');
	var confirmPassword = $('#inputConfirmNewPassword');

	if (password.val().length < 8 || password.val().length > 64 || (passwordRegex.test(password.val()) === false)) {
		addInputDangerClass(password);
		e.preventDefault();
	}
	if (confirmPassword.val() != password.val()) {
		addInputDangerClass(confirmPassword);
		e.preventDefault();
	}

	var formData = new FormData($(this)[0]);
	formData.append('token', getUrlVars()['token']);

	$.ajax({
		url: 'http://localhost:9090/api/auth/passwordchange',
		type: 'POST',
		data: formData,
		success: function (data) {
			addFlashMessage('#passwordChange', 'Congratulations', 'You have successfully changed your password', 'success');
			window.location.replace("/login?token=" + data["token"]);
		},
		error: function (data) {
			addFlashMessage('#passwordChange', data.statusText, data.responseText, 'danger');
		},
		cache: false,
		contentType: false,
		processData: false
	});

});