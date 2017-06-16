var express = require('express');
var http = require('http');
var router = express.Router();

/* GET home page. */
router.get('/', function(req, res, next) {
	if (typeof req.cookies['token'] !== 'undefined' && req.cookies['token'] !== null) {
		res.render('pindex', { title: 'Search for movies.. | Hypertube', page: 'index'});
	} else {
		res.render('index', { title: 'Home | Hypertube' });
	}
});

/* GET log in page. */
router.get('/login', function(req, res, next) {
  res.render('login', { title: 'Log In | Hypertube', page: 'login'});
});

/* GET sign up page. */
router.get('/signup', function(req, res, next) {
  res.render('signup', { title: 'Sign Up | Hypertube', page: 'signup'});
});

/* GET private home page. */
router.get('/media', function(req, res, next) {
	if (req.cookies['token'] != undefined) {
		res.render('media', { title: 'Media | Hypertube', page: 'media'});
	} else {
		res.redirect(401, '/login');
	}
});

/* GET private profile page. */
router.get('/profile', function(req, res, next) {
	if (req.cookies['token'] != undefined) {
		res.render('profile', { title: 'Profile | Hypertube', page: 'profile'});
	} else {
		res.redirect(401, '/login');
	}
});

router.get('/profile/:id', function(req , res) {

	console.log(req.url);

	var options = {
		host: 'localhost',
		port: '9090',
		path: '/api/profile?id=' + req.params.id,
		method: 'GET',
		headers: {
			'Authorization': 'Bearer ' + req.cookies.token,
			'Content-Type': 'application/json'
		}
	};

	var str = "";

	var request = http.request(options, function(response) {

		response.on('data', function (chunk) {
			str += chunk;
		});

		response.on('end', function () {
			console.log(JSON.parse(str));
			res.render('public_profile', { title: 'Profile | Hypertube', page: 'public_profile', user: JSON.parse(str) });
		});

	});

	request.end();

});

/* GET private profile page. */
router.get('/recover', function(req, res, next) {
  res.render('recover', { title: 'Recover | Hypertube', page: 'recover'});
});

/* GET private profile page. */
router.get('/passwordchange', function(req, res, next) {
  res.render('passwordchange', { title: 'Password Change | Hypertube', page: 'passwordchange' });
});

/* GET private profile page. */
router.get('/logout', function(req, res, next) {
	cookie = req.cookies;
	for (var prop in cookie) {
		if (!cookie.hasOwnProperty(prop)) {
			continue;
		}
		res.cookie(prop, '', {expires: new Date(0)});
	}
	res.redirect('/');
});

router.get('/favicon.ico', function(req, res) {
	res.sendStatus(204);
});

module.exports = router;