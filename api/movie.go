package main

import (
	"encoding/json"
	"net/url"
	"net/http"
)

/* 
 * listMovies return a list of medias
 */
func listMovies(page string) (Medias, error) {

	if page == "" {
		page = "1"
	}

	resp, err := http.Get("https://yts.ag/api/v2/list_movies.json?page=" + page)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var a APIResponse
	err = json.NewDecoder(resp.Body).Decode(&a)
	if err != nil {
		return nil, err
	}

	return a.Data.Medias, nil
}

/* 
 * listQueryMovies return a list of medias found with a query
 */
func listQueryMovies(query string, page string) (Medias, error) {
	resp, err := http.Get("https://yts.ag/api/v2/list_movies.json?query_term=" + url.PathEscape(query) + "&page=" + page)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var a APIResponse
	err = json.NewDecoder(resp.Body).Decode(&a)
	if err != nil {
		return nil, err
	}
	return a.Data.Medias, nil
}

/* 
 * getShowDetails return information about a movie
 */
func getMovieDetails(ytsID string, imdbID string) (*APIMedia, error) {

	var a APIResponse
	resp, err := http.Get("https://yts.ag/api/v2/movie_details.json?movie_id=" + ytsID + "&with_images=true&with_cast=true")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&a)
	if err != nil {
		return nil, err
	}

	for key, value := range a.Data.Media.Torrents {
		a.Data.Media.Torrents[key].URL = "magnet:?xt=urn:btih:" + value.Hash + "&dn=" + url.PathEscape(a.Data.Media.Title) + "&tr=udp://open.demonii.com:1337/announce&tr=udp://tracker.openbittorrent.com:80&tr=udp://tracker.coppersurfer.tk:6969&tr=udp://glotorrents.pw:6969/announce&tr=udp://tracker.opentrackr.org:1337/announce&tr=udp://torrent.gresille.org:80/announce&tr=udp://p4p.arenabg.com:1337&tr=udp://tracker.leechers-paradise.org:6969"
	}

	resp, err = http.Get("http://www.omdbapi.com/?i=" + imdbID + "&plot=full")
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&a.Data.Media.Omdb)
	if err != nil {
		return nil, err
	}

	return &a.Data.Media, nil
}