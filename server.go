/**
Server runs the http interface for this short linker.

Author: Danver Braganza
*/

package main

import (
//	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"html/template"

	"github.com/gorilla/mux"
	"github.com/blevesearch/bleve"
)


var index = connectToBleve("links.bleve")

type Shortcut struct {
	Url string
	ShortForm string
	Description string
}

// FindShortcut looks up the given shortcut requested in a map.
func FindShortcut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := bleve.NewQueryStringQuery(vars["shortcut"])
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := index.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), 501)
	}

	print(searchResult.String())

	if len(searchResult.Hits) == 1 {
		http.Redirect(w, r, searchResult.Hits[0].ID, 307)
	}

	if err != nil {

	} else {
		ShowForm(w, r)
	}
}

// NormalizeUrl takes the urls that we get and puts some sane defaults to it.
func NormalizeUrl(s string) string {
	u, _ := url.Parse(s)
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	return u.String()
}


func ShowForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortcut := vars["shortcut"]
	t := template.Must(template.New("index").Parse(`<html><body>
<form action="/" method="POST">
<center>
Url to shorten: <input type="text" name="shortcut" value="{{.}}">
Redirect text <input type="text" name="url">
<input type="submit">
</form></center>`))
	t.Execute(w, shortcut)
}


// SetShortcut will put the first url in the form as a shortcut.
func SetShortcut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortcut, ok := vars["shortcut"]
	r.ParseForm()
	if !ok {
		sc, ok := r.Form["shortcut"]
		if !ok || len(sc) == 0 {
			http.Error(w, "Shortcut was not supplied", 400)
			return
		}
		shortcut = sc[0]
	}

	urls := r.Form["url"]

	if len(urls) > 0 {
		url := NormalizeUrl(urls[0])
		print("Setting ", shortcut, " to ", url, "\n")
		index.Index(url, Shortcut{
			Url: url,
			ShortForm: shortcut,
		})
		http.Redirect(w, r, url, 307)
	} else {
		http.Error(w, "URL was not supplied", 400)
	}
}

func connectToBleve(indexFilePath string) bleve.Index {
	if _, err := os.Stat(indexFilePath); err != nil {
		index, err := bleve.New(indexFilePath, bleve.NewIndexMapping())
		if err != nil {
			log.Fatal(err)
		}
		return index
	 } else {
		 index, err := bleve.Open(indexFilePath)
		 if err != nil {
			 log.Fatal(err)
		 }
		 return index
	 }
}

func main () {

	r := mux.NewRouter()

	r.HandleFunc("/index.html", ShowForm)
	r.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index.html", 307)
	}).Methods("GET")

	r.HandleFunc("/{shortcut}", FindShortcut).Methods("GET")
	r.HandleFunc("/", SetShortcut).Methods("POST")
	r.HandleFunc("/{shortcut}", SetShortcut).Methods("POST")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":80", nil))

}
