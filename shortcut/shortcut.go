//Shortcut contains types and methods for manipulating shortcuts.
package shortcut

import (
	"log"
	"net/url"
	"os"

	"github.com/blevesearch/bleve"
)

//A Shortcut is a mapping from a shortform string to an alternative url.
type Shortcut struct {
	Url         string
	ShortForm   string
	Description string
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

func NewIndex(indexFilePath string) Index {
	if _, err := os.Stat(indexFilePath); err != nil {
		index, err := bleve.New(indexFilePath, bleve.NewIndexMapping())
		if err != nil {
			log.Fatal(err)
		}
		return Index{index}
	} else {
		index, err := bleve.Open(indexFilePath)
		if err != nil {
			log.Fatal(err)
		}
		return Index{index}
	}
}

func (i Index) FindShortcut(query string) (results []string, err error) {
	bleveQuery := bleve.NewQueryStringQuery(query)
	searchRequest := bleve.NewSearchRequest(bleveQuery)
	searchResult, err := i.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	for _, result := range searchResult.Hits {
		results = append(results, result.ID)
	}
	return results, nil
}

func (i Index) SetShortcut(url, shortform string) (normalizedUrl string) {
	normalizedUrl = NormalizeUrl(url)
	i.Index.Index(normalizedUrl, Shortcut{
		Url:       normalizedUrl,
		ShortForm: shortform,
	})
	return
}
