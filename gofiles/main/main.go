package main

import (
	_ "bytes"
	_ "encoding/json"
	"fmt"
	"html/template"
	_ "io/ioutil"
	"net/http"

	"willkuvarti.com/mdgetter"
)

type htmlrequests struct {
	Md        string
	Slidemenu string
}

var event htmlrequests

func gethtmlvar() htmlrequests {
	var ret htmlrequests
	ret.Md = mdgetter.Getfile("html/last-activity.html")
	ret.Slidemenu = mdgetter.Getfile("html/slidemenu.html")
	return ret
}

func index(w http.ResponseWriter, r *http.Request) {
	var dosyam = "html/index.html"
	t, err := template.ParseFiles(dosyam)
	if err != nil {
		fmt.Println("There is an error while page on loading : ", err)
		return
	}
	event = gethtmlvar()
	t.Execute(w, event)
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/test":
		fmt.Fprint(w, "Test Workes")
	default:
		index(w, r)
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./html")))
	//http.HandleFunc("/", handler)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		panic(err)
	}
}
