//Package Shortcut contains types and methods for manipulating shortcuts.
package shortcut

import (
	"log"
	"net/url"
	"os"

	"github.com/blevesearch/bleve"
	// Import the following two analyzers.
	_ "github.com/blevesearch/bleve/analysis/analyzers/keyword_analyzer"
	_ "github.com/blevesearch/bleve/analysis/analyzers/web"
)

//A Shortcut is a mapping from a shortform string to an alternative url.
type Shortcut struct {
	URL,
	ShortForm,
	Description string
}

// Type ensures that Shortcut implements bleve.Classifier
func (Shortcut) Type() string {
	return "Shortcut"
}

// FromFields creates a Shortcut from a dictionary of fields.
func FromFields(fields map[string]interface{}) Shortcut {
	retval := Shortcut{
		URL:       fields["URL"].(string),
		ShortForm: fields["ShortForm"].(string),
	}

	if description, ok := fields["Description"].(string); ok {
		retval.Description = description
	}

	return retval
}

//An Index is a searchable collection of shortcuts.
type Index struct {
	bleve.Index
}

// NormalizeURL takes the urls that we get and puts some sane defaults to it.
func NormalizeURL(s string) string {
	u, _ := url.Parse(s)
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	return u.String()
}

func setUpBleve(indexFilePath string) (bleve.Index, error) {
	shortcutMapping := bleve.NewDocumentMapping()

	shortFormFieldMapping := bleve.NewTextFieldMapping()
	shortFormFieldMapping.Analyzer = "keyword"
	shortcutMapping.AddFieldMappingsAt("ShortForm", shortFormFieldMapping)

	descriptionFieldMapping := bleve.NewTextFieldMapping()
	descriptionFieldMapping.Analyzer = "web"
	descriptionFieldMapping.IncludeTermVectors = true
	shortcutMapping.AddFieldMappingsAt("Description", descriptionFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("Shortcut", shortcutMapping)

	return bleve.New(indexFilePath, indexMapping)
}

// NewIndex creates a new Shotcut.Index, initialized from a file.
func NewIndex(indexFilePath string) Index {
	if _, err := os.Stat(indexFilePath); err != nil {
		index, err := setUpBleve(indexFilePath)

		if err != nil {
			log.Fatal(err)
		}
		return Index{index}
	}

	index, err := bleve.Open(indexFilePath)
	if err != nil {
		log.Fatal(err)
	}
	return Index{index}
}

// AddShortcut adds a given shortcut to this index, and returns the normalized URL string.
func (i Index) AddShortcut(s Shortcut) (normalizedURL string) {
	normalizedURL = NormalizeURL(s.URL)
	s.URL = normalizedURL
	i.Index.Index(s.ShortForm, s)
	return
}

// FindShortcut attempts to find a given string by first searching for an
// absolute match. If that doesn't exist, it will perform a partial match on all
// the text. Results is a list of 0, 1 or more shortcuts that were found. If
// sole is true, that result was an absolute match.
func (i Index) FindShortcut(query string) (results []Shortcut, sole bool, err error) {
	sole = true
	termQ := bleve.NewTermQuery(query).SetField("ShortForm")
	termSR := bleve.NewSearchRequest(termQ)
	termSR.Fields = []string{"URL", "ShortForm", "Description"}
	searchResult, err := i.Search(termSR)
	if err != nil {
		return nil, false, err
	}
	if len(searchResult.Hits) == 0 {
		sole = false
		// Didn't find anything. Let's widen our search.
		q := bleve.NewQueryStringQuery(query).SetField("*")
		matchSR := bleve.NewSearchRequest(q)
		matchSR.Fields = []string{"URL", "ShortForm", "Description"}
		searchResult, err = i.Search(matchSR)
		if err != nil {
			return nil, false, err
		}
		log.Print(searchResult.Hits)
	}

	for _, result := range searchResult.Hits {
		results = append(results, FromFields(result.Fields))
	}
	return results, sole, nil
}
