package render

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// sequence generates a sequence of integers from 1 to the specified number passed to template
func sequence(n int) []int {
	seq := make([]int, n)
	for i := range seq {
		seq[i] = i + 1
	}
	return seq
}

// RenderTemplate writes the data to browser via templates
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	// create a template cache
	tc, err := createTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	// get requested template from cache
	t, ok := tc[tmpl]
	if !ok {
		log.Fatal(err)
	}

	// buf := new(bytes.Buffer)

	err = t.Execute(w, data)
	if err != nil {
		log.Println(err)
	}

	// // render the template
	// _, err = buf.WriteTo(w)
	// if err != nil {
	// 	log.Println(err)
	// }
	return nil
}

func createTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// get all of the files named *.page.tmpl from ./templates
	pages, err := filepath.Glob("./templates/*.html")
	if err != nil {
		return myCache, err
	}

	// range through all files ending with *.page.tmpl
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(template.FuncMap{"sequence": sequence}).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
