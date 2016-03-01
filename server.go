/*
Server runs the http interface for this link shortener.

Author: Danver Braganza
*/

package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/danverbraganza/shortlink/fetcher"
	"github.com/danverbraganza/shortlink/shortcut"
)

// ShortcutHandler is a struct that handles all requests.
type ShortcutHandler struct {
	index        shortcut.Index
	formTemplate *template.Template
}

// Get looks up the given shortcut requested in the index.
func (s ShortcutHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	query := vars["shortform"]

	results, sole, err := s.index.FindShortcut(query)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sole {
		http.Redirect(w, r, results[0].URL, http.StatusSeeOther)
		return
	}
	s.ShowForm(w, r, results)
}

// ShowForm shows a nice form where the user can enter a new url.
func (s ShortcutHandler) ShowForm(w http.ResponseWriter, r *http.Request, partMatch []shortcut.Shortcut) {
	vars := mux.Vars(r)
	sf := vars["shortform"]
	s.formTemplate.Execute(w, struct {
		ShortForm string
		Shortcuts []shortcut.Shortcut
	}{sf, partMatch})
}

// ShowEmptyForm specifically shows an empty form.
func (s ShortcutHandler) ShowEmptyForm(w http.ResponseWriter, r *http.Request) {
	s.ShowForm(w, r, nil)
}

// Post saves the "first" url found as the "first" shortform found.
// You may pass shortforms as form params or in the url.
func (s ShortcutHandler) Post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	r.ParseForm()

	shortform, ok := vars["shortform"]

	if !ok {
		// Shortform not provided in url: find it in the form.
		sc := r.Form["shortform"]
		if len(sc) == 0 {
			http.Error(w, "Shortcut was not supplied", http.StatusBadRequest)
			return
		}
		shortform = sc[0]
	}

	if _, sole, _ := s.index.FindShortcut(shortform); sole {
		http.Error(w, "Shortcut already exists.", http.StatusBadRequest)
		return
	}

	urls := r.Form["url"]
	if len(urls) == 0 {
		http.Error(w, "URL was not supplied", http.StatusBadRequest)
		return
	}
	url := urls[0]

	var description string
	if descriptions, ok := r.Form["description"]; ok {
		description = descriptions[0]
	}
	normalizedURL := shortcut.NormalizeURL(url)
	http.Redirect(w, r, normalizedURL, http.StatusSeeOther)
	go func() {
		if description == "" && r.Form["attempt"] != nil {
			description = fetcher.FindDescription(normalizedURL)
		}
		log.Print("Setting ", shortform, " to ", normalizedURL, ": ", description)
		s.index.AddShortcut(shortcut.Shortcut{
			ShortForm:   shortform,
			URL:         normalizedURL,
			Description: description,
		})
	}()
}

func main() {
	handler := ShortcutHandler{
		shortcut.NewIndex("links.bleve"),
		template.Must(template.ParseFiles("templates/form.tmpl")),
	}
	r := mux.NewRouter()

	r.HandleFunc("/index.html", handler.ShowEmptyForm)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
	}).Methods("GET")

	r.HandleFunc("/{shortform:.*}", handler.Get).Methods("GET")
	r.HandleFunc("/", handler.Post).Methods("POST")
	r.HandleFunc("/{shortform:.*}", handler.Post).Methods("POST")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
