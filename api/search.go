package main

import (
	"strconv"
	"net/http"
	"encoding/json"
	"log"
	"strings"
	"net/url"
)

type APIQueryResponse struct {
	Genre		string	`json:"genre"`
	Rating		string	`json:"imdbRating"`
	Year		string	`json:"year"`
}

func sortByGenres(genreChosen string, list Medias) (Medias) {

	var newList Medias
	for _, listvalue := range list {
		resp, err := http.Get("http://www.omdbapi.com/?i=" + listvalue.IMDBID + "&plot=short")
		if err != nil {
			log.Println("Error call omdbAPI")
			return nil
		}
		var a APIQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&a)
		if err != nil {
			log.Println("Error Decode json")
			return nil
		}
		if strings.Contains(a.Genre, genreChosen) {
			var show APIMedia
			show.Title = listvalue.Title
			show.ID = listvalue.ID
			show.Year = listvalue.Year
			show.CoverURL = listvalue.CoverURL
			show.Summary = listvalue.Summary
			show.IMDBID = listvalue.IMDBID
			newList = append(newList, show)
		}
	}
	return newList
}

func sortByRating(ratingChosen string, list Medias) (Medias) {

	ratingChos, err := strconv.Atoi(ratingChosen)
	if err != nil {
		log.Println("Error to convert ratingChosen: Atoi")
	}
	var newList Medias
	for _, listvalue := range list {
		resp, err := http.Get("http://www.omdbapi.com/?i=" + listvalue.IMDBID + "&plot=short")
		if err != nil {
			log.Println("Error call omdbAPI")
			return nil
		}
		var a APIQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&a)
		if err != nil {
			log.Println("Error Decode json")
			return nil
		}

		ratingMedia, err := strconv.ParseFloat(a.Rating, 64)
		if err != nil {
			log.Println("Error to convert ratingMedia")
		}

		if int(ratingMedia) >= ratingChos*2 {
			var show APIMedia
			show.Title = listvalue.Title
			show.ID = listvalue.ID
			show.Year = listvalue.Year
			show.CoverURL = listvalue.CoverURL
			show.Summary = listvalue.Summary
			show.IMDBID = listvalue.IMDBID
			newList = append(newList, show)
		}
	}
	return newList
}

func sortByDateFrom(dateFromChosen string, list Medias) (Medias) {

	dateFromChos, err := strconv.Atoi(dateFromChosen)
	if err != nil {
		log.Println("Error to convert dateFromChosen: Atoi")
	}
	var newList Medias
	for _, listvalue := range list {
		resp, err := http.Get("http://www.omdbapi.com/?i=" + listvalue.IMDBID + "&plot=short")
		if err != nil {
			log.Println("Error call omdbAPI")
			return nil
		}
		var a APIQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&a)
		if err != nil {
			log.Println("Error Decode json")
			return nil
		}
		omdbYear, err := strconv.Atoi(a.Year[:4])
		if err != nil {
			log.Println("Error to convert omdbYearx	: Atoi")
		}
		if omdbYear >= dateFromChos {
			var show APIMedia
			show.Title = listvalue.Title
			show.ID = listvalue.ID
			show.Year = listvalue.Year
			show.CoverURL = listvalue.CoverURL
			show.Summary = listvalue.Summary
			show.IMDBID = listvalue.IMDBID
			newList = append(newList, show)
		}
	}
	return newList
}

func sortByDateTo(dateToChosen string, list Medias) (Medias) {

	dateToChos, err := strconv.Atoi(dateToChosen)
	if err != nil {
		log.Println("Error to convert dateToChosen: Atoi")
	}
	var newList Medias
	for _, listvalue := range list {
		resp, err := http.Get("http://www.omdbapi.com/?i=" + listvalue.IMDBID + "&plot=short")
		if err != nil {
			log.Println("Error call omdbAPI")
			return nil
		}
		var a APIQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&a)
		if err != nil {
			log.Println("Error Decode json")
			return nil
		}

		omdbYear, err := strconv.Atoi(a.Year[:4])
		if err != nil {
			log.Println("Error to convert omdbYearx	: Atoi")
		}

		if omdbYear <= dateToChos {
			var show APIMedia
			show.Title = listvalue.Title
			show.ID = listvalue.ID
			show.Year = listvalue.Year
			show.CoverURL = listvalue.CoverURL
			show.Summary = listvalue.Summary
			show.IMDBID = listvalue.IMDBID
			newList = append(newList, show)
		}
	}
	return newList
}

func sortByFilters(urlQuery url.Values, list Medias) (Medias) {

	if urlQuery.Get("genres") != "" {
		list = sortByGenres(urlQuery.Get("genres"), list)
	}
	if urlQuery.Get("rating") != "" {
		list = sortByRating(urlQuery.Get("rating"), list)
	}
	if urlQuery.Get("dateFrom") != "" {
		list = sortByDateFrom(urlQuery.Get("dateFrom"), list)
	}
	if urlQuery.Get("dateTo") != "" {
		list = sortByDateTo(urlQuery.Get("dateTo"), list)
	}
	return list
}