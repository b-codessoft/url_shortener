package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	urlshort "github.com/b-codessoft/url_shortener"
)

func main() {
	yamlFileName := flag.String("yamlFile", "paths.yml", "a yaml file in the format 'path, url'")
	jsonFileName := flag.String("jsonFile", "paths.json", "a json file in the format '{path: url}'")
	flag.Parse()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	// Build the YAMLHandler using the mapHandler as the
	// fallback
	yamlData, err := os.ReadFile(*yamlFileName)
	if err != nil {
		panic(err)
	}

	yamlHandler, err := urlshort.YAMLHandler(yamlData, mapHandler)
	if err != nil {
		panic(err)
	}

	jsonData, err := os.ReadFile(*jsonFileName)
	if err != nil {
		panic(err)
	}

	jsonHandler, err := urlshort.JSONHandler(jsonData, yamlHandler)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", jsonHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
