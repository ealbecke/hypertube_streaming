package main

import (
	"math/rand"
	"encoding/base64"
	"net/http"
	"io"
	"os"
	"mime"
	"image/png"
	_ "image/jpeg"
	"encoding/json"
	"net/smtp"
	"image"
	"path/filepath"
	"log"
	"github.com/nfnt/resize"
)

/* 
 * saveFormFile
 */
func saveFormFile(fileInputName string, path string, w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile(fileInputName)
	if err == http.ErrMissingFile {
		http.Error(w, "File is missing", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Println("[Hypertube] - saveFormFile(): Error: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		log.Println("[Hypertube] - saveFormFile(): Error: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	file.Seek(0, 0)
	if mime.TypeByExtension(filepath.Ext(fileHeader.Filename)) != http.DetectContentType(buffer) {
		log.Println("[Hypertube] - saveFormFile(): File type is not supported")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f, err := os.OpenFile(path + ".png", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("[Hypertube] - saveFormFile(): Error with os.OpenFile() call: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Println("[Hypertube] - saveFormFile(): Could not decode image with image.Decode() call:" + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	m := resize.Resize(200, 267, img, resize.Lanczos3)

	err = png.Encode(f, m)
	if err != nil {
		log.Println("[Hypertube] - saveFormFile(): Could not decode image with png.Encode() call:" + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = createThumbnail(m, path)
	if err != nil {
		log.Println("[Hypertube] - saveFormFile(): Could not decode image with createThumbnail() call:" + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

/* 
 * randToken
 */
func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

/* 
 * sendEmail
 */
func sendEmail(to string, message string) {
	auth := smtp.PlainAuth("", "hypertubeealbecke@gmail.com", "Hypertube123", "smtp.gmail.com")
	msg := []byte(message)
	err := smtp.SendMail("smtp.gmail.com:25", auth, "hypertubeealbecke@gmail.com", []string{to}, msg)
	if err != nil {
		log.Println("[Hypertube] - sendEmail(): Error with smtp.SendMail() call: " + err.Error())
		return
	}
}

/* 
 * createThumbnail
 */
func createThumbnail(img image.Image, path string) error {
	fileThumbnail, err := os.Create(path + "-thumbnail.png")
	if err != nil {
		return err
	}
	defer fileThumbnail.Close()
	m := resize.Resize(64, 64, img, resize.Lanczos3)
	png.Encode(fileThumbnail, m)
	if err != nil {
		return err
	}

	return nil
}

/* 
 * saveImageFromURL
 */
func saveImageFromURL(url string, path string) error {
	file, err := os.Create(path + ".png")
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	file.Seek(0, 0)

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	file.Seek(0, 0)
	m := resize.Resize(200, 267, img, resize.Lanczos3)
	err = png.Encode(file, m)
	if err != nil {
		return err
	}

	err = createThumbnail(m, path)
	if err != nil {
		return err
	}

	return nil
}

/* 
 * writeJSON
 */
func writeJSON(w http.ResponseWriter, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Println("[Hypertube] - writeJSON(): " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("content-type", "application/json; charset=utf-8")
	_, err = w.Write(data)
	if err != nil {
		log.Println("[Hypertube] - writeJSON(): " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

/* 
 * removeOldMedias
 */
func removeOldMedias() {
	log.Println("GoCron: run every days to check and delete medias older than a month")
	rows, err := db.Query("SELECT torrent_name FROM torrents WHERE creation_date < DATE_SUB(NOW(), INTERVAL 1 MONTH)")
	if err != nil {
		log.Print(err.Error())
	}
	for rows.Next() {
		var torrentName string
		err := rows.Scan(&torrentName)
		if err != nil {
			log.Println("[Hypertube] - removeOldMedias(): SQL Error: " + err.Error())
			return
		}
		err = os.RemoveAll("downloads/" + torrentName)
		if err != nil {
			log.Println("[Hypertube] - removeOldMedias(): Error in os.RemoveAll() call: " + err.Error())
			return
		}
		stmt, err := db.Prepare("DELETE FROM torrents WHERE torrent_name = ?")
		if err != nil {
			log.Println("[Hypertube] - removeOldMedias(): SQL Error: " + err.Error())
		}
		_, err = stmt.Exec(torrentName)
		if err != nil {
			log.Println("[Hypertube] - removeOldMedias(): SQL Error: " + err.Error())
		}
	}
}