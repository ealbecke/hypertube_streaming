package main

import (
	"github.com/odwrtw/eztv"
	"strconv"
	"net/http"
	"encoding/json"
	"strings"
)

/* 
 * listShows return a list of medias
 */
func listShows(page string) (Medias, error) {

	intpage, err := strconv.Atoi(page)
	if err != nil {
		return nil, err
	}

	list, err := eztv.ListShows(intpage)
	if err != nil {
		return nil, err
	}

	var shows Medias
	for _, value := range list {
		var show APIMedia
		show.Title = value.Title
		show.Year, _ = strconv.Atoi(value.Year)
		show.CoverURL = value.Images.Poster
		show.Summary = value.Synopsis
		show.IMDBID = value.ImdbID
		show.Show = true
		shows = append(shows, show)
	}

	return shows, nil
}

/* 
 * listQueryShow return a list of medias found with a query
 */
func listQueryShow(query string) (Medias, error) {

	list, err := eztv.SearchShow(query)
	if err != nil {
		return nil, err
	}

	var shows Medias
	for _, value := range list {
		var show APIMedia
		show.Title = value.Title
		show.Year, _ = strconv.Atoi(value.Year)
		show.CoverURL = value.Images.Poster
		show.Summary = value.Synopsis
		show.IMDBID = value.ImdbID
		show.Show = true
		shows = append(shows, show)
	}
	return shows, nil
}

/* 
 * getShowDetails return information about one show
 */
func getShowDetails(id string) (*APIMedia, error) {

	resp, err := eztv.GetShowDetails(id)
	if err != nil {
		return nil, err
	}

	var show APIMedia
	show.Title = resp.Title
	show.Year, _ = strconv.Atoi(resp.Year)
	show.CoverURL = resp.Images.Poster
	show.Summary = resp.Synopsis
	show.ID, _ = strconv.Atoi(resp.ID)
	show.IMDBID = resp.ImdbID

	for _, value := range resp.Episodes {
		var torrent APITorrent
		torrent.Title = value.Title
		for key, val := range value.Torrents {
			if value.Torrents[key] != nil {
				if !strings.Contains(val.URL, ".mkv") {
					torrent.Peers = val.Peers
					torrent.URL = val.URL
					torrent.Seeds = val.Seeds
					show.Torrents = append(show.Torrents, torrent)
				}
			}
		}
	}

	ret, err := http.Get("http://www.omdbapi.com/?i=" + id + "&plot=full")
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(ret.Body).Decode(&show.Omdb)
	if err != nil {
		return nil, err
	}

	return &show, nil
}