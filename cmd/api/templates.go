package main

import (
	"github.com/justinas/nosurf"
	"greenlight.abdulalsh.com/internal/data"
	"greenlight.abdulalsh.com/ui"

	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"
)

type templateData struct {
	CurrentYear int
	Message     string
	CSRFToken   string
	User        data.User
}

// we create a newTemplateData() helper which will return a pointer to a templatedata struct initilized with the current year,
// note that we are not using the http.request param, but we will do later
func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear: time.Now().Year(),
		CSRFToken:   nosurf.Token(r),
	}
}
func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}

	// Use the filepath.Glob() function to get a slice of all filepaths that
	// match the pattern "./ui/html/pages/*.tmpl". This will essentially gives
	// us a slice of all the filepaths for our application 'page' templates
	// like: [ui/html/pages/home.tmpl ui/html/pages/view.tmpl]
	pages, err := fs.Glob(ui.Files, "html/pages/*.gohtml")
	if err != nil {
		return nil, err
	}
	// Loop through the page filepaths one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.gohtml') from the full filepath
		// and assign it to the name variable.
		name := filepath.Base(page)
		// Parse the base template file into a template set.
		//the func map must be registered with the template set before we call the parse files
		//this means we have to use templates.New to create an empty template set, use the funcs() method to register the template func map then parse the files as normal

		patterns := []string{
			"html/base.gohtml",
			//"html/partials/*.gohtml",
			page,
		}
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		// Add the template set to the map, using the name of the page
		// (like 'home.gohtml') as the key.
		// Add the template set to the map as normal...

		cache[name] = ts
	}
	// Return the map.
	return cache, nil

}

// creating custome functions for templates
func humanDate(t time.Time) string {
	// Return the empty string if time has the zero value.
	if t.IsZero() {
		return ""
	}

	// Convert the time to UTC before formatting it.
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// we will init a template.FuncMap object and store it in a global variable. this is essentially a string-keyed map which acts as a lookup
// between the names of our custom template functions and the function itself
var functions = template.FuncMap{
	"humanDate": humanDate,
}
