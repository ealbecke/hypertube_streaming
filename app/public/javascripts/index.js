$(function() {

	getUser(displayUserInfoHeader);

	var movieList;

	var page = 1;

	getListMovies(page++);

	var delay = (function() {
		var timer = 0;
		return function(callback, ms) {
			clearTimeout (timer);
			timer = setTimeout(callback, ms);
		};
	})();

	var searchForm = {
		'Search': "",
		'Genres': "",
		'Rating': "",
		'DateFrom': "",
		'DateTo': ""
	};

	$('#inputSearchGenres').change(function() {
		$( "#movieList" ).empty();
		page = 1;
		searchForm.Genres = getSelectedValue($(this));
		searchMedias(searchForm, page++)
	});

	$('#inputSearchRatings').change(function() {
		$( "#movieList" ).empty();
		page = 1;
		searchForm.Rating = getSelectedValue($(this));
		searchMedias(searchForm, page++)
	});

	$('#inputSearchDateFrom').change(function() {
		$( "#movieList" ).empty();
		page = 1;
		searchForm.DateFrom = getSelectedValue($(this));
		searchMedias(searchForm, page++)
	});

	$('#inputSearchDateTo').change(function() {
		$( "#movieList" ).empty();
		page = 1;
		searchForm.DateTo = getSelectedValue($(this));
		searchMedias(searchForm, page++)
	});

	$('#inputSearch').keyup(function() {
		$( "#movieList" ).empty();
		page = 1;
		delay(function() {
			searchForm.Search = $('#inputSearch').val()
			if (searchForm.Search != "") {
				enableQueryButtons()
				searchMedias(searchForm, page++)
			} else {
				disableQueryButtons()
				$( "#movieList" ).empty();
				getListMovies(page++)
			}
		}, 500);
	});

	$(window).scroll(function() {
		if ($(document).height() - $(window).height() == $(window).scrollTop()) {
			 if (searchForm.Search) {
			 	searchMedias(searchForm, page++);
			 } else {
			 	getListMovies(page++);
			 }
		}
	});

});

function disableQueryButtons() {
	$('#inputSearchGenres').prop('disabled', true);
	$('#inputSearchRatings').prop('disabled', true);
	$('#inputSearchDateFrom').prop('disabled', true);
	$('#inputSearchDateTo').prop('disabled', true);
}

function enableQueryButtons() {
	$('#inputSearchGenres').prop('disabled', false);
	$('#inputSearchRatings').prop('disabled', false);
	$('#inputSearchDateFrom').prop('disabled', false);
	$('#inputSearchDateTo').prop('disabled', false);
}

function printDelayMessage() {
	addFeedbackMessage('#inputSearchOptions', 'There is no films for your search...')
}

function searchMedias(searchQuery, page) {
	if (searchQuery.Search) {
		var url = 'http://localhost:9090/api/get_list_medias?q=' + searchQuery.Search;
		

		if (searchQuery.Genres) {
			url += '&genres=' + searchQuery.Genres;
		}

		if (searchQuery.Rating) {
			url += '&rating=' + searchQuery.Rating;
		}

		if (searchQuery.DateFrom) {
			url += '&dateFrom=' + searchQuery.DateFrom;
		}

		if (searchQuery.DateTo) {
			url += '&dateTo=' + searchQuery.DateTo;
		}

		url += '&page=' + page;

		$.ajax({
			type: 'GET',
			headers: {
				'Authorization':'Bearer ' + Cookies.get('token')
			},
			url: url,
			dataType: 'json',
			success: function(data) {
				removeFeedbackMessage('#inputSearchOptions')
				 if (data == null) {
					setTimeout(printDelayMessage, 1000)
					removeFeedbackMessage('#inputSearchOptions')


				 } else {
					addMoviesToList(data);
				}
			},
			error: function (data) {
				console.log(data)
			}
		});
	} else {
		$( "#movieList" ).empty();
		page = 1;
		getListMovies(page++);
	}
}

function getListMovies(page) {
	$.ajax({
		type: 'GET',
		headers: {
			'Authorization':'Bearer ' + Cookies.get('token')
		},
		url: 'http://localhost:9090/api/get_list_medias?page=' + page,
		dataType: 'json',
		success: function(data) {
			addMoviesToList(data);
		}
	});
}

function addMoviesToList(data) {
	var movieTemplate = '';
	$.each(data, function(index, value) {
		rating = Math.ceil(value.rating/2);

		var ratingTemplate = "";
		var i = 0;
		if (value.rating != 0)
		{
			while (i < 5) {
				if (rating) {
					ratingTemplate += '<i class="fa fa-star" aria-hidden="true"></i>'
					rating--;
				} else {
					ratingTemplate += '<i class="fa fa-star-o" aria-hidden="false"></i>'
				}
				i++;
			}
		}

		movieTemplate = '<div class="col-md-6 col-lg-4 col-xl-2 mb-3">';
		
		if (value["already_seen"]) {
			movieTemplate += '<div class="card card-outline-info">';
		} else {
			movieTemplate += '<div class="card">';
		}

		if (value["Show"]) {
			movieTemplate += '<a href=\'' + '/media?show_id=' + value["id"] + '&imdbid=' + value["imdb_code"] + '\'>';
		} else {
			movieTemplate += '<a href=\'' + '/media?movie_id=' + value["id"] + '&imdbid=' + value["imdb_code"] + '\'>';
		}
		if (value["Show"]) {
			movieTemplate += '<h3 class="card-header text-center">' + value.title  + ' <i class=\'fa fa-television\' aria-hidden=\'true\'></i></h3>';
		} else {
			movieTemplate += '<h3 class="card-header text-center">' + value.title + ' <i class=\'fa fa-film\' aria-hidden=\'true\'></i></h3>';
		}

		movieTemplate +=	'<div class="cover-image-container">' +	
								'<img src="' + value.large_cover_image + '" class="card-img-top img-fluid hidden-sm-down">' +
							'</div>' + 
							'</a>' +
							'<ul class="list-group list-group-flush">' +
								'<li class="list-group-item justify-content-center">';
		movieTemplate += ratingTemplate;
		movieTemplate += '</li></ul></div></div>';
		$('#movieList').append(movieTemplate);
	});
}