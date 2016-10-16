package main

import (
	"bytes"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

var upload_directory = ""
var templates map[string]*template.Template

func handleStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	s, ok := binData["s/"+path]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")

	w.Write([]byte(s))
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if p := recover(); p != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			debug.PrintStack()
		}
	}()

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ff, fh, err := r.FormFile("file")
	if err != nil {
		log.Panicln(err)
	}

	filename := ""
	{
		t := time.Now()
		filename = t.Format("2016-01-02_15-04-05_") + fh.Filename
	}
	log.Println("New file:", filename)

	fb, err := ioutil.ReadAll(ff)
	if err != nil {
		log.Panicln(err)
	}

	err = ioutil.WriteFile(filepath.Join(upload_directory, filename), fb, 0644)
	if err != nil {
		log.Panicln(err)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("/skins/"))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if p := recover(); p != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			debug.PrintStack()
		}
	}()

	var buf bytes.Buffer

	err := templates["index"].Execute(&buf, nil)
	if err != nil {
		log.Panicln(err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.Write(buf.Bytes())
}

func initTemplates() {
	it := func(filenames ...string) *template.Template {
		t := template.New("")

		for _, f := range filenames {
			template.Must(t.Parse(binData[f]))
		}
		return t
	}

	templates = make(map[string]*template.Template)

	templates["index"] = it("base.html", "index.html")
}

func initHandlers() {
	ih := func(pattern string, f func(http.ResponseWriter, *http.Request)) {
		h := http.HandlerFunc(f)

		http.Handle(pattern, http.StripPrefix(pattern, h))
	}

	ih("/", handleRoot)
	ih("/s/", handleStatic)
	ih("/new", handleNew)
}

func main() {
	_listen := flag.String("l", "127.0.0.1:8080", "address to listen on")
	_directory := flag.String("d", "/tmp/laseruploads", "directory for storing uploaded files")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)

	upload_directory = *_directory

	if err := os.MkdirAll(upload_directory, 0755); err != nil {
		log.Fatalln(err)
	}

	initTemplates()
	initHandlers()

	log.Printf("Listening on http://%v/", *_listen)
	log.Fatal(http.ListenAndServe(*_listen, nil))
}
