package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

var indexPage *template.Template = nil
var linkPage *template.Template = nil

type TokenContainer struct {
	Token string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func loadPage(title string) []byte {
	filename := title
	file, err := ioutil.ReadFile(filename)
	check(err)
	return file
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if indexPage == nil {
		indexPage, _ = template.ParseFiles("index.html")
	}
	log.Printf("Page :: %s %s", r.Method, r.URL)
	indexPage.Execute(w, struct{}{})
}

type Location struct {
	Lat       float32
	Lon       float32
	Accuracy  float32
	Timestamp int
}

var locations = map[string]Location{}

var locationRegExp, _ = regexp.Compile("^\\/l\\/([a-zA-Z0-9+\\/]+={0,2})$")

func linkHandler(w http.ResponseWriter, r *http.Request) {
	if linkPage == nil {
		linkPage, _ = template.ParseFiles("link.html")
	}
	log.Printf("Page :: %s %s", r.Method, r.URL)

	matches := locationRegExp.FindStringSubmatch(r.URL.Path)
	if len(matches) != 2 {
		fmt.Fprintf(w, "Location link is invalid.\n")
		return
	}

	if token, ok := locations[matches[1]]; ok {
		linkPage.Execute(w, token)
	} else {
		fmt.Fprintf(w, "Location link not found. It might have expired?\n")
	}

}

func newLocationId() string {
	f, err := os.Open("/dev/urandom")
	check(err)
	defer f.Close()

	r := make([]byte, 20)
	f.Read(r)

	id := base64.StdEncoding.EncodeToString(r)
	return id
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("API  :: %s %s", r.Method, r.URL)

	decoder := json.NewDecoder(r.Body)
	var location Location
	err := decoder.Decode(&location)
	check(err)

	id := newLocationId()
	locations[id] = location
	token := &TokenContainer{Token: id}
	response, err := json.Marshal(token)
	check(err)
	fmt.Fprintf(w, string(response))
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/l/", linkHandler)
	http.HandleFunc("/api/v1/request", apiHandler)
	http.ListenAndServe(":1123", nil)
}
