/**
Server runs the http interface for this short linker.

Author: Danver Braganza
*/

package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/danverbraganza/shortlink/shortcut"
)

type ShortcutHandler struct {
	index shortcut.Index
}

// Get looks up the given shortcut requested in the index.
func (s ShortcutHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	query := vars["shortcut"]

	results, err := s.index.FindShortcut(query)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Print(results)

	if len(results) == 1 {
		http.Redirect(w, r, results[0], http.StatusSeeOther)
	} else {
		// TODO(danver): More than one result. Do something clever.
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		ShowForm(w, r)
	}
}

func ShowForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortcut := vars["shortcut"]
	t := template.Must(template.New("index").Parse(`<html><body>
<form action="/" method="POST">
<center>
Url to shorten: <input type="text" name="shortform" value="{{.}}">
Redirect text <input type="text" name="url">
<input type="submit">
</form></center>`))
	t.Execute(w, shortcut)
}

// Posting will save the "first" url found as the "first" shortform found.
// You may pass shortforms as form params or in the url.
func (s ShortcutHandler) Post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortform, ok := vars["shortform"]
	r.ParseForm()
	if !ok {
		sc, ok := r.Form["shortform"]
		if !ok || len(sc) == 0 {
			http.Error(w, "Shortcut was not supplied", http.StatusBadRequest)
			return
		}
		shortform = sc[0]
	}

	urls := r.Form["url"]

	if len(urls) > 0 {
		url := urls[0]
		log.Print("Setting ", shortform, " to ", url, "\n")
		normalizedUrl := s.index.SetShortcut(url, shortform)
		http.Redirect(w, r, normalizedUrl, http.StatusSeeOther)
	} else {
		http.Error(w, "URL was not supplied", http.StatusBadRequest)
	}
}

func main() {
	handler := ShortcutHandler{shortcut.NewIndex("links.bleve")}
	r := mux.NewRouter()

	r.HandleFunc("/index.html", ShowForm)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
	}).Methods("GET")

	r.HandleFunc("/{shortcut}", handler.Get).Methods("GET")
	r.HandleFunc("/", handler.Post).Methods("POST")
	r.HandleFunc("/{shortcut}", handler.Post).Methods("POST")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
