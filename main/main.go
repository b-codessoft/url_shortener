package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	urlshort "github.com/b-codessoft/url_shortener"
	bolt "go.etcd.io/bbolt"
)

const bucketName string = "UrlShortener"

func main() {
	yamlFileName := flag.String("yamlFile", "paths.yml", "a yaml file in the format 'path, url'")
	jsonFileName := flag.String("jsonFile", "paths.json", "a json file in the format '{path: url}'")
	flag.Parse()

	db, err := bolt.Open("url.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	// pathsToUrls := map[string]string{
	// 	"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
	// 	"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	// }
	// mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	// NEW: DB handler as base fallback
	dbHandler := urlshort.DBHandler(db, bucketName, mux)

	// Build the YAMLHandler using the mapHandler as the
	// fallback
	yamlData, err := os.ReadFile(*yamlFileName)
	if err != nil {
		log.Fatal(err)
	}

	yamlHandler, err := urlshort.YAMLHandler(yamlData, db, bucketName, dbHandler)
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := os.ReadFile(*jsonFileName)
	if err != nil {
		log.Fatal(err)
	}

	jsonHandler, err := urlshort.JSONHandler(jsonData, db, bucketName, yamlHandler)
	if err != nil {
		log.Fatal(err)
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
