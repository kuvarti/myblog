package mdgetter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func Getfile(fname string) string {
	content, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Print("some Error Accuares : ", err)
	}
	return (string(content))
}

func PostRequest() string {
	const githubAPI_URL = "https://api.github.com/markdown/raw"

	requestBody := strings.NewReader(Getfile("deneme.md"))
	res, err := http.Post(githubAPI_URL, "text/plain", requestBody)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	content, _ := ioutil.ReadAll(res.Body)
	return (string(content))
}
