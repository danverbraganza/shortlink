//Shortcut contains types and methods for manipulating shortcuts.
package shortcut

import (
	"log"
	"net/url"
	"os"

	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/analysis/analyzers/keyword_analyzer"
)

//A Shortcut is a mapping from a shortform string to an alternative url.
type Shortcut struct {
	Url,
	ShortForm,
	Description string
}

// Shortcut implements bleve.Classifier
func (Shortcut) Type() string {
	return "Shortcut"
}

func FromFields(fields map[string]interface{}) Shortcut {
	retval := Shortcut{
		Url:       fields["Url"].(string),
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

// NormalizeUrl takes the urls that we get and puts some sane defaults to it.
func NormalizeUrl(s string) string {
	u, _ := url.Parse(s)
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	return u.String()
}

func SetUpBleve(indexFilePath string) (bleve.Index, error) {
	shortcutMapping := bleve.NewDocumentMapping()

	shortFormFieldMapping := bleve.NewTextFieldMapping()
	shortFormFieldMapping.Analyzer = "keyword"
	shortcutMapping.AddFieldMappingsAt("ShortForm", shortFormFieldMapping)

	descriptionFieldMapping := bleve.NewTextFieldMapping()
	descriptionFieldMapping.Analyzer = "en"
	descriptionFieldMapping.IncludeTermVectors = true
	shortcutMapping.AddFieldMappingsAt("Description", descriptionFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("Shortcut", shortcutMapping)

	return bleve.New(indexFilePath, indexMapping)
}

func NewIndex(indexFilePath string) Index {
	if _, err := os.Stat(indexFilePath); err != nil {
		index, err := SetUpBleve(indexFilePath)

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

func (i Index) SetShortcut(url, shortform, description string) (normalizedUrl string) {
	normalizedUrl = NormalizeUrl(url)
	i.Index.Index(normalizedUrl, Shortcut{
		Url:         normalizedUrl,
		ShortForm:   shortform,
		Description: description,
	})
	return
}

// FindShortcut attempts to find a given string by first searching for an
// absolute match. If that doesn't exist, it will perform a partial match on all
// the text. Results is a list of 0, 1 or more shortcuts that were found. If
// sole is true, that result was an absolute match.
func (i Index) FindShortcut(query string) (results []Shortcut, sole bool, err error) {
	log.Print(query)
	sole = true
	termQ := bleve.NewTermQuery(query).SetField("ShortForm")
	termSR := bleve.NewSearchRequest(termQ)
	termSR.Fields = []string{"Url", "ShortForm", "Description"}
	searchResult, err := i.Search(termSR)
	if err != nil {
		return nil, false, err
	}
	if len(searchResult.Hits) == 0 {
		sole = false
		// Didn't find anything. Let's widen our search.
		q := bleve.NewMatchPhraseQuery(query) //.SetField("*")
		matchSR := bleve.NewSearchRequest(q)
		matchSR.Fields = []string{"Url", "ShortForm", "Description"}
		searchResult, err = i.Search(matchSR)
		if err != nil {
			return nil, false, err
		}
	}

	for _, result := range searchResult.Hits {
		results = append(results, FromFields(result.Fields))
	}
	return results, sole, nil
}
