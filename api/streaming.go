package main

import (
	"database/sql"
	"net/http"
	"strings"
	"html"
	"io"
	"log"
	"path/filepath"
	"time"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

/* 
 * getDownloadMediaHandler start the torrent download and send a reponse at 5%
 * of the download. If the torrent is already downloaded, it respond as soon as it can.
 */
func getDownloadMediaHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	
		magnetURL := html.UnescapeString(r.URL.Query().Get("magnet"))
		imdbid := r.URL.Query().Get("imdbid")

		m, err := metainfo.ParseMagnetURI(magnetURL)
		if err != nil {
			log.Println("[Hypertube] - getDownloadMediaHandler(): Error could not parse margnet url: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		token, err := tokenParse(getAuthorizationToken(r.Header))
		if err != nil {
			log.Println("[Hypertube] - getDownloadMediaHandler(): Token error: " + err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

			stmt, err := db.Prepare("INSERT INTO medias_seen(user_id, imdb_id) VALUES(?, ?)")
			if err != nil {
				log.Println("[Hypertube] - getDownloadMediaHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = stmt.Exec(claims["id"].(float64), imdbid)
			if err != nil {
				log.Println("[Hypertube] - getDownloadMediaHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		}

		var ID int64
		var extension string
		err = db.QueryRow("SELECT id, extension FROM torrents WHERE hash = ?", m.InfoHash.HexString()).Scan(&ID, &extension)
		if err == sql.ErrNoRows {

			var t *torrent.Torrent
			if t, err = c.AddMagnet(magnetURL); err != nil {
				log.Println("[Hypertube] - getDownloadMediaHandler(): Error adding magnet url to torrent client: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			<-t.GotInfo()
			t.DownloadAll()

			file := getLargestFile(t)
			file.PrioritizeRegion(0, int64(t.NumPieces()/100*5))

			stmt, err := db.Prepare("INSERT INTO torrents(torrent_name, hash, extension) VALUES(?, ?, ?)")
			if err != nil {
				log.Println("[Hypertube] - getDownloadMediaHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			res, err := stmt.Exec(file.Path(), m.InfoHash.HexString(), strings.TrimPrefix(filepath.Ext(file.Path()), "."))
			if err != nil {
				log.Println("[Hypertube] - getDownloadMediaHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			ID, err = res.LastInsertId()
			if err != nil {
				log.Println("[Hypertube] - getDownloadMediaHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			go func() {
				for {
					<-time.After(2 * time.Second)
					if percentage(t) < 100 {
						log.Println("[Hypertube] - getDownloadMediaHandler(): " + m.DisplayName + ": " + strconv.Itoa(int(percentage(t))))
					} else {
						return
					}
				}
			}()

			for {
				<-time.After(2 * time.Second)
				if percentage(t) > 5 {
					idString := strconv.FormatInt(ID, 10)
					data := struct 
					{
						ID string `json:"torrent_id"`
						Extension string `json:"extension"`
					}{ idString, strings.TrimPrefix(filepath.Ext(file.Path()), ".") }
					writeJSON(w, data)
					return
				}
			}
			
		} else if err != nil {
			log.Println("[Hypertube] - getDownloadMediaHandler(): SQL Error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		stmt, err := db.Prepare("UPDATE torrents SET creation_date = NOW() WHERE id = ?")
		if err != nil {
			log.Println("[Hypertube] - getDownloadMediaHandler(): SQL Error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(ID)
		if err != nil {
			log.Println("[Hypertube] - getDownloadMediaHandler(): SQL Error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		idString := strconv.FormatInt(ID, 10)
					data := struct 
					{
						ID string `json:"torrent_id"`
						Extension string `json:"extension"`
					}{ idString, extension }
		writeJSON(w, data)
		return
	}
}

/* 
 * getLargestFile get the largest file in a torrent content
 */
func getLargestFile(t *torrent.Torrent) *torrent.File {
	var target torrent.File
	var maxSize int64

	for _, file := range t.Files() {
		if maxSize < file.Length() {
			maxSize = file.Length()
			target = file
		}
	}
	return &target
}

/* 
 * percentage return the downloaded percentage of a torrent
 */
func percentage(t *torrent.Torrent) float64 {
	info := t.Info()
	if info == nil {
		return 0
	}
	return float64(t.BytesCompleted()) / float64(info.TotalLength()) * 100
}

type SeekableContent interface {
	io.ReadSeeker
	io.Closer
}

type FileEntry struct {
	*torrent.File
	*torrent.Reader
}

/* 
 * Seek move the cursor on the media in order to play it where the user want
 */
func (f FileEntry) Seek(offset int64, whence int) (int64, error) {
	return f.Reader.Seek(offset+f.File.Offset(), whence)
}

/* 
 * NewFileReader return a SeekableContent which is a new reader on the media
 */
func NewFileReader(f *torrent.File) (SeekableContent, error) {
	torrent := f.Torrent()
	reader := torrent.NewReader()

	reader.SetReadahead(f.Length() / 100)
	reader.SetResponsive()
	_, err := reader.Seek(f.Offset(), io.SeekStart)

	return &FileEntry{
		File:   f,
		Reader: reader,
	}, err
}

/* 
 * getMediaStreamHandler return a media stream in order to play the media
 */
func getMediaStreamHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		torrentID := r.URL.Query().Get("torrent_id")

		var torrentName string
		var hash string
		err := db.QueryRow("SELECT torrent_name, hash FROM torrents WHERE id = ?", torrentID).Scan(&torrentName, &hash)
		if err == sql.ErrNoRows {
			log.Println("[Hypertube] - getMediaStreamHandler(): SQL Error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if file, ok := c.Torrent(metainfo.NewHashFromHex(hash)); ok {

			target := getLargestFile(file)
			entry, err := NewFileReader(target)
			if err != nil {
				log.Println("[Hypertube] - getMediaStreamHandler(): Error opening a new file reader")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer func() {
				if err := entry.Close(); err != nil {
					log.Println("[Hypertube] - getMediaStreamHandler(): Error closing the file reader")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}()

			flusher := w.(http.Flusher)
			flusher.Flush()
			w.Header().Set("Content-Disposition", "attachment; filename=\"" + torrentName + "\"")
			http.ServeContent(w, r, target.Path(), time.Now(), entry)
		} else {
			log.Println("[Hypertube] - getMediaStreamHandler(): Can't find torrent in torrent client")
			w.WriteHeader(http.StatusNotFound)
			return
		}

	}
}