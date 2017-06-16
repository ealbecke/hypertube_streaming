var currentRequest = null;

$(function() {

	getUser(displayUserInfoHeader);

	var players = plyr.setup({ blankUrl: "https://cdn.selz.com/plyr/blank.mp4" });

	overlay = $('<div data-toggle="tooltip" data-placement="top" title="You must choose a stream to start" data-animation="false">').prependTo('.movie-player').attr('id', 'overlay');
	loading = $('<div class="sk-fading-circle"><div class="sk-circle1 sk-circle"></div><div class="sk-circle2 sk-circle"></div><div class="sk-circle3 sk-circle"></div><div class="sk-circle4 sk-circle"></div><div class="sk-circle5 sk-circle"></div><div class="sk-circle6 sk-circle"></div><div class="sk-circle7 sk-circle"></div><div class="sk-circle8 sk-circle"></div><div class="sk-circle9 sk-circle"></div><div class="sk-circle10 sk-circle"></div><div class="sk-circle11 sk-circle"></div><div class="sk-circle12 sk-circle"></div></div>')
	
	$('#overlay').tooltip()

	createMedia(getUrlVars()["imdbid"])

	$('#sendButtonMessage').click(function(e) {
		postComment(e);
	});

});

function createMedia(imdb_id) {
	$.ajax({
		url: "http://localhost:9090/api/create_media",
		type: 'POST',
		contentType: 'application/json',
		data: JSON.stringify({"imdbid": imdb_id}),
		headers: {
			'Authorization':'Bearer ' + Cookies.get('token')
		}, 
		success: function () {
			getMovie(displayMediaInfo);	
		},
		error: function () {
			window.location.replace("/");
		},
	});
}

function postComment(e) {
	e.preventDefault();

	var message = $('#commentCntn');
	var submit = true;

	if (message.val().length < 1 || message.val().length > 250  || !specialCharRegex.test(message.val())) {
		addInputDangerClass(message);
		addFeedbackMessage(message, 'Comment must be Alphanumeric and must be under 250 characters');
		submit = false;
	}

	if (submit === true) {
		var commentObj = {"comment": $('#commentCntn').val(), "mediaImdbId": getUrlVars()['imdbid']};

		$.ajax({
			url: "http://localhost:9090/api/create_message",
			type: 'POST',
			contentType: 'application/json',
			data: JSON.stringify(commentObj),
			headers: {
				'Authorization':'Bearer ' + Cookies.get('token')
			},
			success: function (data) {
				var commentTemplate =	'<li class=\'list-group-item media\'>' +
											'<img class=\'d-flex mr-3 rounded-circle img-thumbnail align-self-start hidden-sm-down\', src=\'/images/profiles/' + data.Username + data.AuthorID + '-thumbnail.png' + '\'>' +
											'<div class=\'media-body\'>' + 
												'<h5 class=\'mt-0 mb-1\'><a href=\'/profile\'>' + data.Username + '</a></h5>' +
												'<p>' +  data.Message + '</p>' +
											'</div>' + 
										'</li>';
				$('#commentCntn').val('');
				$(".formcomment").after(commentTemplate);
				removeFeedbackMessage('#movieComments')
				removeInputDangerClass('#commentCntn')
			},
			error: function (data) {
				addFlashMessage('#movieComments', data.statusText, data.responseText, 'danger');
			},
		});
	}
}

function getMovie(handleData) {

	var api_url = ""
	if (getUrlVars()["movie_id"] != undefined) {
		api_url = "http://localhost:9090/api/get_media_details?movie_id=" + getUrlVars()["movie_id"] + "&imdbid=" + getUrlVars()["imdbid"];
	} else if (getUrlVars()["show_id"] != undefined) {
		api_url = "http://localhost:9090/api/get_media_details?show_id=" + getUrlVars()["show_id"] + "&imdbid=" + getUrlVars()["imdbid"];
	} else {
		window.location.replace("/");
	}

	$.ajax({
		url: api_url,
		type: 'GET',
		headers: {
				'Authorization':'Bearer ' + Cookies.get('token')
			},
		success: function (data) {
			handleData(data);
		},
		error: function (data) {
			addFlashMessage('.movie-player', data.statusText, data.responseText, 'danger');
		},
	});
}

function getMediaStream(elem, subtitles, data) {

	var button = elem;
	loading.remove();
	overlay.remove();
	button.removeClass("btn-info");
	button.addClass("btn-success");

	var players = plyr.get('.movie-player');

	if ($.isEmptyObject(subtitles) != true) {
		players[0].source({
			type:       'video',
			title:      'Example title',
			sources: [{
				src:    'http://localhost:9090/api/get_media_stream?torrent_id=' + data.torrent_id,
				type:   'video/' + data.extension
			}],
			tracks:	[{
				kind:   	'captions',
				label:		'English00',
				srclang:	'en',
				src:		subtitles[0],
				default: 	true
			}]
		});
	} else {
		players[0].source({
			type:       'video',
			title:      'Example title',
			sources: [{
				src:    'http://localhost:9090/api/get_media_stream?torrent_id=' + data.torrent_id,
				type:   'video/' + data.extension
			}]
		});
	}

	players[0].play();
}

function startDownloadCall(elem) {
		
	var subtitles = {};
	var button = elem;
	var ajaxURL = 'http://localhost:9090/api/start_download_media?imdbid=' + getUrlVars()['imdbid'] + '&magnet=' + encodeURIComponent(button.data('magnet-url'))

	currentRequest = $.ajax({
		url: ajaxURL,
		type: 'GET',
		dataType: 'json',
		headers: {
			'Authorization':'Bearer ' + Cookies.get('token')
		},
		beforeSend: function() {

			$('#overlay').tooltip('dispose')
			loading.prependTo('.movie-player').attr('id', 'loading');
			scroll(0,0);

			if (currentRequest != null) {
				currentRequest.abort();
			}

		},
		success: function (data) {

			if (getUrlVars()['movie_id'] != undefined) {
				$.ajax({
					url: 'http://localhost:9090/api/get_media_subtitles?movie=true&imdb_id=' + getUrlVars()['imdbid'],
					type: 'GET',
					dataType: 'json',
					headers: {
						'Authorization':'Bearer ' + Cookies.get('token')
					},
					success: function (dataSub) {
						if (dataSub != null) {
							subtitles = dataSub;
						}
						getMediaStream(elem, subtitles, data);
					}
				});
			} else {
				getMediaStream(elem, subtitles, data);
			}
		},
		error: function (data) {
			button.removeClass("btn-info");
			button.addClass("btn-danger");
		},
	});

}

function displayMediaInfo(data) {
	$('.movie-player video').attr('poster', data['background_image_original']);
	$('#moviePicture').attr('src', data['large_cover_image']);
	$('#movieTitle').find('h2').text(data['title']);
	$('#moviePicture').attr('src', data['large_cover_image']);
	$('#synopsis').after(data.omdbinfos['plot']);
	$('#director').after(data.omdbinfos['director']);
	$('#starring').after(data.omdbinfos['actors']);
	$('#releaseDate').after(data.omdbinfos['released']);
	$('#genres').after(data.omdbinfos['genre']);
	$('#runningTime').after(data.omdbinfos['runtime']);

	$.each(data["torrents"], function(index, value) {

		var streamTemplate = '<li class=\'list-group-item d-block\'>';

		if (value['url'] != undefined) {
			streamTemplate += '<button type=\'button\' class=\'btn btn-secondary launchTorrent\' data-magnet-url=\'' + value['url'] + '\'>Torrent: ' + value['title'] + '</button>';
		};
		if (value['quality'] != undefined) {
			streamTemplate += ' Quality: ' + value['quality'];
		};
		if (value['size'] != undefined) {
			streamTemplate += ' - Size: ' + value['size'];
		};
		if (value['seeds'] != undefined) {
			streamTemplate += ' - Seeds: ' + value['seeds'];
		};
		if (value['peers'] != undefined) {
			streamTemplate += ' - Peers: ' + value['peers'];
		};

		streamTemplate += '</li>';
		$('#streams').append(streamTemplate);

	});

	$("#streams .launchTorrent").on("click", function() {
		startDownloadCall($(this));
	});

	$.each(data["comments"], function(index, value) {
		var commentTemplate = '<li class=\'list-group-item media\'>' +
		'<img class=\'d-flex mr-3 rounded-circle img-thumbnail align-self-start hidden-sm-down\', src=\'/images/profiles/' + value.Username + value.AuthorID + '-thumbnail.png' + '\'>' +
		'<div class=\'media-body\'>' + 
		'<h5 class=\'mt-0 mb-1\'><a href=\'/profile/' + value.AuthorID + ' \'>' + value.Username + '</a></h5>' +
		'<p>' + value.Message + '</p>' +
		'</div></li>';

		$('#movieComments').append(commentTemplate);
	});
}

