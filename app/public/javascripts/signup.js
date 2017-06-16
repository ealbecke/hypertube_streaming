$(function() {

	if (Cookies.get('token') != undefined) {
		Cookies.remove('token');
		location.reload();
	}

	bindInputStates('#inputSignUpEmail', 8, 200, emailRegex, messageErrorEmailRegExp, messageErrorEmailSize);
	bindInputStates('#inputSignUpUsername', 8, 15, specialCharRegex, messageErrorUsernameRegExp, messageErrorUsernameSize);
	bindInputStates('#inputSignUpFirstname', 1, 25, specialCharRegex, messageErrorFirstnameRegExp, messageErrorFirstnameSize);
	bindInputStates('#inputSignUpLastname', 1, 25, specialCharRegex, messageErrorLastnameRegExp, messageErrorLastnameSize);
	bindInputStates('#inputSignUpPassword', 8, 64, passwordRegex, messageErrorPasswordRegExp, messageErrorPasswordSize);
	bindFileInputStates('#inputSignUpProfilePicture');

	$('#inputSignUpConfirmPassword').focusin(function(e) {
		removeFeedbackMessage(this);
		$(this).parent().removeClass('has-success');
		$(this).removeClass('form-control-success');
		$(this).parent().removeClass('has-danger');
		$(this).removeClass('form-control-danger');
	});

	$('#inputSignUpConfirmPassword').focusout(function(e) {
		if ($(this).val().length != 0 && $(this).val() === $('#inputSignUpPassword').val()) {
			addInputSuccessClass(this);
		} else {
			addInputDangerClass(this);
			addFeedbackMessage(this, messageErrorConfirmPassword);
		}
	});

});

$('#signupForm').submit(function(e) {
	e.preventDefault();

	var email = $('#inputSignUpEmail');
	var username = $('#inputSignUpUsername');
	var profilePicture = $('#inputSignUpProfilePicture');
	var firstname = $('#inputSignUpFirstname');
	var lastname = $('#inputSignUpLastname');
	var password = $('#inputSignUpPassword');
	var confirmPassword = $('#inputSignUpConfirmPassword');

	var submit = true;

	if (email.val().length <= 8 || email.val().length >= 200 || emailRegex.test(email.val()) === false) {
		addInputDangerClass(email);
		submit = false;
	}

	if (username.val().length < 8 || username.val().length > 15 || specialCharRegex.test(username.val()) === false) {
		addInputDangerClass(username);
		submit = false;
	}

	if ($.inArray(profilePicture.val().split('.').pop().toLowerCase(), ['jpg', 'jpeg', 'png']) == -1) {
		addInputDangerClass(profilePicture);
		submit = false;
	}	

	if (firstname.val().length <= 1 || firstname.val().length >= 25  || specialCharRegex.test(firstname.val()) === false) {
		addInputDangerClass(firstname);
		submit = false;
	}

	if (lastname.val().length <= 1 || lastname.val().length >= 25  || specialCharRegex.test(lastname.val()) === false) {
		addInputDangerClass(lastname);
		submit = false;
	}

	if (password.val().length <= 8 || password.val().length >= 64 || passwordRegex.test(password.val())=== false) {
		addInputDangerClass(password);
		submit = false;
	}

	if (confirmPassword.val() != password.val() || !confirmPassword.val()) {
		addInputDangerClass(confirmPassword);
		submit = false;
	}

	if (submit === true) {

		var formData = new FormData($(this)[0]);

		$.ajax({
			url: 'http://localhost:9090/api/auth/register',
			type: 'POST',
			data: formData,
			success: function (data) {
				addFlashMessage('#signupForm', 'Congratulations', 'You have successfully registered', 'success');
				window.location.replace("/login?token=" + data["token"]);
			},
			error: function (data) {
				addFlashMessage('#signupForm', data.statusText, data.responseText, 'danger');
			},
			cache: false,
			contentType: false,
			processData: false
		});

	}
});

$('#inputLogInGoogle').click(function() {
	window.location.replace('http://localhost:9090/api/auth/google');
});

$('#inputLogIn42').click(function() {
	window.location.replace('http://localhost:9090/api/auth/fortytwo');
});