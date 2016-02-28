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
	index        shortcut.Index
	formTemplate *template.Template
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

	switch len(results) {
	case 0:
		s.ShowForm(w, r)
	case 1:
		http.Redirect(w, r, results[0], http.StatusSeeOther)
	default:
		// TODO(danver): Handle the multiple case.
		s.ShowForm(w, r)
	}
}

// ShowForm shows a nice form where the user can enter a new url.
func (s ShortcutHandler) ShowForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortcut := vars["shortcut"]
	s.formTemplate.Execute(w, struct {
		Shortcut string
	}{shortcut})
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

	urls := r.Form["url"]
	if len(urls) == 0 {
		http.Error(w, "URL was not supplied", http.StatusBadRequest)
		return
	}

	url := urls[0]
	log.Print("Setting ", shortform, " to ", url, "\n")
	normalizedUrl := s.index.SetShortcut(url, shortform)
	http.Redirect(w, r, normalizedUrl, http.StatusSeeOther)
}

func main() {
	handler := ShortcutHandler{
		shortcut.NewIndex("links.bleve"),
		template.Must(template.ParseFiles("templates/form.tmpl")),
	}
	r := mux.NewRouter()

	r.HandleFunc("/index.html", handler.ShowForm)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
	}).Methods("GET")

	r.HandleFunc("/{shortcut}", handler.Get).Methods("GET")
	r.HandleFunc("/", handler.Post).Methods("POST")
	r.HandleFunc("/{shortcut}", handler.Post).Methods("POST")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
