$(function() {

	getUser(displayUserInfoHeader);
	getUser(displayUserInfo);


	bindInputStates('#inputUpdateUsername', 8, 15, specialCharRegex, messageErrorUsernameRegExp, messageErrorUsernameSize);
	bindInputStates('#inputUpdateFirstname', 1, 25, specialCharRegex, messageErrorFirstnameRegExp, messageErrorFirstnameSize);
	bindInputStates('#inputUpdateLastname', 1, 25, specialCharRegex, messageErrorLastnameRegExp, messageErrorLastnameSize);
	bindInputStates('#inputUpdateEmail', 8, 200, emailRegex, messageErrorEmailRegExp, messageErrorEmailSize);
	bindFileInputStates('#inputUpdateProfilePicture');

	$('#inputUpdateLanguage').focusout(function(e) {
		addInputSuccessClass(this);
	});

	$('#inputUpdateOldPassword').focusout(function(e) {
		removeFeedbackMessage(this)
		removeInputDangerClass(this);
		if ($(this).val().length > 0) {
			if ($(this).val().length < 8 || $(this).val().length > 64 || (passwordRegex.test($(this).val()) === false)) {
				addFeedbackMessage(this, messageErrorPasswordSize + ', '+ messageErrorPasswordRegExp);
				addInputDangerClass(this);
			}
		}
	});

	$('#inputUpdateNewPassword').focusout(function(e) {
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

	$('#inputUpdateConfirmNewPassword').focusout(function(e) {
		removeFeedbackMessage(this)
		removeInputDangerClass(this);
		removeInputSuccessClass(this);
		if ($(this).val().length > 0) {
			if ($(this).val() != $('#inputUpdateNewPassword').val()) {
				addInputDangerClass(this);
				addFeedbackMessage(this, messageErrorConfirmPassword);
			} else {
				addInputSuccessClass(this);
			}
		}
	});

});

function displayUserInfo(data) {
	$('#profilePicture').attr('src', '/images/profiles/' + data['Username'] + data['ID'] + '.png')
	$('#profile').find('h2').text(data['Username'])
	$("#profileUsername").text(data['Firstname'] + ' ' + data['Lastname']);
	$("#profileMemberSince").text(data['Creation_date'].slice(0, 10));
	$("#profileNumberComments").text(data['CommentsNumber']);
	$("#profileNumberWatchedMovie").text(data['MoviesSeenNumber']);
	$("#profilePreferedGenres").text('Drama');

	$('#inputUpdateUsername').val(data['Username']);
	$('#inputUpdateFirstname').val(data['Firstname']);
	$('#inputUpdateLastname').val(data['Lastname']);
	$('#inputUpdateEmail').val(data['Email']);
	$('#inputUpdateLanguage').val(data['Language_id']);

	if (data['Provider_id'] != '1') {
		$('#inputUpdateOldPassword').parent().remove();
		$('#inputUpdateNewPassword').parent().remove();
		$('#inputUpdateConfirmNewPassword').parent().remove();
	}
}

$('#updateProfileForm').submit(function(e) {
	e.preventDefault();

	var username = $('#inputUpdateUsername');
	var firstname = $('#inputUpdateFirstname');
	var lastname = $('#inputUpdateLastname');
	var email = $('#inputUpdateEmail');
	var language = $('#inputUpdateLanguage');
	var profilePicture = $('#inputUpdateProfilePicture');
	var oldPassword = $('#inputUpdateOldPassword');
	var newPassword = $('#inputUpdateNewPassword');
	var confirmNewPassword = $('#inputUpdateConfirmNewPassword');

	var submit = true;

	if (username.val().length < 8 || username.val().length > 15 || specialCharRegex.test(username.val()) === false) {
		addInputDangerClass(username);
		submit = false;
	}

	if (firstname.val().length < 1 || firstname.val().length > 25  ||
		!specialCharRegex.test(firstname.val())) {
		addInputDangerClass(firstname);
		submit = false;
	}

	if (lastname.val().length < 1 || lastname.val().length > 25  ||
		!specialCharRegex.test(lastname.val())) {
		addInputDangerClass(lastname);
		submit = false;
	}

	if (email.val().length < 8 || email.val().length > 200 || 
		!emailRegex.test(email.val())) {
		addInputDangerClass(email);
		submit = false;
	}

	if (profilePicture.val()) {
		if ($.inArray(profilePicture.val().split('.').pop().toLowerCase(), ['jpg', 'jpeg', 'png']) == -1) {
			addInputDangerClass(profilePicture);
			submit = false;
		}
	}

	if (oldPassword.val() && newPassword.val() && confirmNewPassword.val()) {
		if (oldPassword.val().length > 0 || newPassword.val().length > 0 || confirmNewPassword.val().length > 0) {
			if (oldPassword.val().length < 8 || oldPassword.val().length > 64 || passwordRegex.test(oldPassword.val()) === false) {
				addInputDangerClass(oldPassword);
				submit = false;
			}
			if (newPassword.val().length < 8 || newPassword.val().length > 64 || (passwordRegex.test(newPassword.val()) === false)) {
				addInputDangerClass(newPassword);
				submit = false;
			}
			if (confirmNewPassword.val().length < 8 || confirmNewPassword.val().length > 64 || (passwordRegex.test(confirmNewPassword.val()) === false)) {
				addInputDangerClass(confirmNewPassword);
				submit = false;
			}
		} else {
			removeInputDangerClass(oldPassword);
			removeInputDangerClass(newPassword);
			removeInputDangerClass(confirmNewPassword);
		}
	}

	if (submit === true) {

		var formData = new FormData($(this)[0]);

		$.ajax({
			url: 'http://localhost:9090/api/auth/edit',
			type: 'POST',
			data: formData,
			success: function (data) {
				addFlashMessage('#updateProfileForm', 'Congratulations', 'You have successfully updated your account', 'success');
				location.reload();
			},
			error: function (data) {
				addFlashMessage('#updateProfileForm', data.statusText, data.responseText, 'danger');
			},
			headers: {
				'Authorization':'Bearer ' + Cookies.get('token')
			},
			cache: false,
			contentType: false,
			processData: false
		});

	}
});