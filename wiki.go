package main 

import (
		"log"
		"io/ioutil"
		"net/http"
		"html/template"
		"regexp"
		"errors"
	)

type Page struct {
	Title string
	Body []byte
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html", "tmpl/home.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var link = regexp.MustCompile("\\[([a-zA-Z]+)\\]")

func (p Page) save() error {
	filename := "data/" + p.Title + ".txt"
	body,_ := p.parse()
	return ioutil.WriteFile(filename, body, 0600)
}

func (p Page) parse() ([]byte, error){
	body := link.ReplaceAllFunc(p.Body, func(s []byte) []byte{
		group :=link.ReplaceAllString(string(s), `$1`)
		newGroup := "<a href='/view" + group + "'>" + group + "</a>"
		return []byte(newGroup)
	})
	log.Println(body)
	return body,nil
}

func loadPage(title string) (Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return Page{Title: "", Body:[]byte("")}, err
	}
	return Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, errLoad := loadPage(title)
	if errLoad != nil {
		http.Redirect(w, r, "/edit/"+title, 301)
		return;
	}
	renderTemplate(w,p,"view.html")
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, errL := loadPage(title)
	if errL != nil {
		p = Page{Title: title}
	}
	renderTemplate(w,p,"edit.html")
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	r.ParseForm()
	data := r.Form
	if data != nil {
		body := data.Get("body")
		p := &Page{Title: title, Body: []byte(body)}
		p.save();
		http.Redirect(w, r, "/view/"+title, http.StatusFound)
		return;
	}
	log.Fatal("error reading body")
}

func renderTemplate(w http.ResponseWriter, p Page, file string) {
	err :=  templates.ExecuteTemplate(w, file, p)
	if err != nil {
		log.Fatal("errror rendering")
	}
}

func renderHome(w http.ResponseWriter) {
	err := templates.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
		log.Fatal("error rendering")
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil // The title is the second subexpression.
}

func makeHandler( fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		if r.URL.Path == "/view/" {
			renderHome(w)
			return
		}
		m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8082", nil))
}