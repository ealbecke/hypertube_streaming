var emailRegex = new RegExp('^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+.[A-Za-z]{2,4}$');
var specialCharRegex = new RegExp('^[a-zA-Z0-9\']+(?:[ -]?[a-zA-Z0-9\']+)*$');
var passwordRegex = new RegExp('^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#\$%\^&\*])(?=.{8,})');
var messageErrorEmailSize = 'Email must be between 8 and 200 characters';
var messageErrorEmailRegExp = 'Email address must be valid';
var messageErrorUsernameSize = 'Username must be between 8 and 15 characters';
var messageErrorUsernameRegExp = 'Username must contain only alphanumeric characters';
var messageErrorFirstnameSize = 'Firstname must be between 1 and 25 characters';
var messageErrorFirstnameRegExp = 'Firstname must contain only alphanumeric characters';
var messageErrorLastnameSize = 'Lastname must be between 1 and 25 characters';
var messageErrorLastnameRegExp = 'Lastname must contain only alphanumeric characters';
var messageErrorPasswordSize = 'Password must be between 8 and 64 characters';
var messageErrorPasswordRegExp = 'Password must contain at least 1 uppercase letter, 1 figure and 1 special character';
var messageErrorConfirmPassword = 'Password confirmation doesn\'t match with password';
var messageErrorPictureExtensionRegExp = 'Your picture must be a jpeg, jpg or png';

$(function() {
	if (Cookies.get('token') != undefined) {
		$('.navbar-nav #profileLink').css('display', 'inline-block');
		$('.navbar-nav #logoutLink').css('display', 'inline-block');
		$('.navbar-nav #loginLink').css('display', 'none');
		$('.navbar-nav #signUpLink').css('display', 'none');
	} else {
		$('.navbar-nav #profileLink').css('display', 'none');
		$('.navbar-nav #logoutLink').css('display', 'none');
		$('.navbar-nav #loginLink').css('display', 'inline-block');
		$('.navbar-nav #signUpLink').css('display', 'inline-block');
	}
})

function isLogged() {
	if (Cookies.get('token') != undefined) {
		return false;
	} else {
		return true;
	}
}

function mustBeLogged() {
	if (isLogged() === false) {
		window.location.replace("/login")
	}
}

function getUser(handleData) {
	$.ajax({
		url: 'http://localhost:9090/api/user',
		type: 'GET',
		headers: {
			'Authorization':'Bearer ' + Cookies.get('token')
		},
		success: function (data) {
			handleData(data); 
		}
	});
}

function displayUserInfoHeader(data) {
	$('#profileLinkContent').text(data['Username']);
}

function getUrlVars() {
	var vars = [], hash;
	var hashes = window.location.href.slice(window.location.href.indexOf('?') + 1).split('&');
	for(var i = 0; i < hashes.length; i++)
	{
		hash = hashes[i].split('=');
		vars.push(hash[0]);
		vars[hash[0]] = hash[1];
	}
	return vars;
}

function mustBeLogged() {
	$.ajax({
		url: 'http://localhost:9090/api/user',
		type: 'GET',
		error: function (data) {
			window.location.replace("/login")
		}
	});
}

function bindInputStates(elem, min, max, regExp, messageErrorRegExp, messageErrorSize) {

	$(elem).focusin(function(e) {
		removeFeedbackMessage(elem);
		removeInputSuccessClass(elem);
		removeInputDangerClass(elem);
	});

	$(elem).focusout(function(e) {
		if ($(this).val().length < min || $(this).val().length > max) {
			if (regExp.test($(this).val()) === false) {
				addFeedbackMessage(elem, messageErrorSize + ', '+ messageErrorRegExp);
			} else {
				addFeedbackMessage(elem, messageErrorSize);
			}
			addInputDangerClass(elem);
		} else if (regExp.test($(this).val()) === false) {
			addFeedbackMessage(elem, messageErrorRegExp);
			addInputDangerClass(elem);
		} else {
			addInputSuccessClass(elem);
		}
	});

}

function bindFileInputStates(elem) {

	$(elem).focusin(function(e) {
		removeFeedbackMessage(elem);
		removeInputSuccessClass(elem);
		removeInputDangerClass(elem);
	});

	$(elem).change(function(e) {
		var ext = $(elem).val().split('.').pop().toLowerCase();
		if (ext == 'jpeg' || ext == 'jpg' || ext == 'png') {
			addInputSuccessClass(elem);
		} else {
			addFeedbackMessage(elem, messageErrorPictureExtensionRegExp);
			addInputDangerClass(elem);
		}
	});

}

function addInputDangerClass(elem) {
	$(elem).parent().addClass('has-danger');
	$(elem).addClass('form-control-danger');
}

function addInputSuccessClass(elem) {
	$(elem).parent().addClass('has-success');
	$(elem).addClass('form-control-success');
}

function removeInputDangerClass(elem) {
	$(elem).parent().removeClass('has-danger');
	$(elem).removeClass('form-control-danger');
}

function removeInputSuccessClass(elem) {
	$(elem).parent().removeClass('has-success');
	$(elem).removeClass('form-control-success');
}

function addFlashMessage(elem, title, message, type) {
	$('div').remove('.alert');
	var content = "<div class='alert alert-" + type + " role='alert'>" +
	"<button type='button' class='close' data-dismiss='alert' aria-label='Close'><span aria-hidden='true'>&times;</span></button>" +
	"<h5>" + title + "</h5> " + message + "</div>";
	$(elem + ' :first-child').first().before(content);
}

function addFeedbackMessage(elem, message) {
	removeFeedbackMessage(elem);
	var content = "<div class='form-control-feedback'>"+ message +"</div>"
	setTimeout(function () {
		$(elem).after(content)
	}, 50);
}

function removeFeedbackMessage(elem) {
	$(elem).parent().find('.form-control-feedback').remove();
}

function getSelectedValue(elem) {
	if ($(elem).val() != '') {
		return $(elem).val();
	}
	return null;
}