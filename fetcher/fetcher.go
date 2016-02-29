package fetcher

import (
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// FindDescription attempts to find a string that suitably describes a link.
// First the link is retrieved, then the HTML is inspected for descriptive information.
func FindDescription(url string) string {
	for i := 0; i < 3; i++ {
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		tks := html.NewTokenizer(resp.Body)

		ds := []string{}
		for tks.Next() != html.ErrorToken {
			tk := tks.Token()
			switch tk.DataAtom {
			case atom.Title:
				tks.Next()
				ds = append(ds, tks.Token().Data)
			case atom.Meta:
				var content, name string
				for _, att := range tk.Attr {
					switch att.Key {
					case "content":
						content = att.Val
					case "name":
						name = att.Val
					}
				}
				if content != "" && name != "" {
					switch name {
					case "author", "description", "keywords", "creator":
						ds = append(ds, content)
					}

				}

			}

		}
		return strings.Join(ds, "\n")
	}

	// Tried 3 times, give up and use and empty document.
	return ""
}
