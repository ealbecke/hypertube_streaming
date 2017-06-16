$('#recoverForm').submit(function(e) {
	e.preventDefault();
	
	var email = $('#inputRecoverEmail');

	if (email.val().length < 8 || email.val().length > 200 || emailRegex.test(email.val()) === false) {
		addInputDangerClass(email);
		e.preventDefault();
	}

		var formData = new FormData($(this)[0]);

		$.ajax({
			url: 'http://localhost:9090/api/auth/recover',
			type: 'POST',
			data: formData,
			success: function (data) {
				addFlashMessage('#recoverForm', 'Congratulations', 'Check your email', 'success');
			},
			error: function (data) {
				addFlashMessage('#recoverForm', data.statusText, data.responseText, 'danger');
			},
			cache: false,
			contentType: false,
			processData: false
		});

});
