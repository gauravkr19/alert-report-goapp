package render

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

// RenderTemplate writes the data to browser via templates
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	// create a template cache
	tc, err := CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	// get requested template from cache
	t, ok := tc[tmpl]
	if !ok {
		log.Fatal(err)
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// get all of the files named *.html from ./templates
	pages, err := filepath.Glob("./templates/*.html")
	if err != nil {
		return myCache, err
	}

	// range through all files ending with *.html
	for _, page := range pages {
		name := filepath.Base(page)

		// register custom func and parse
		ts, err := template.New(name).Funcs(template.FuncMap{
			"sequence":   sequence,
			"formatDate": formatDate,
		}).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob("./templates/*.layout.html")
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.html")
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}

// custom function sequence generates a series of integers from 1
func sequence(n int) []int {
	seq := make([]int, n)
	for i := range seq {
		seq[i] = i + 1
	}
	return seq
}

// custom function formatDate to format the date
func formatDate(t *time.Time) string {
	if t == nil {
		return "" // Return empty string for nil values
	}
	return t.Format("2006-01-02 15:04:05.999")
}
