package url_shortener

import (
	"encoding/json"
	"net/http"

	bolt "go.etcd.io/bbolt"
	"gopkg.in/yaml.v3"
)

func DBHandler(db *bolt.DB, bucketName string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		var redirectURL string
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			if b == nil {
				return nil
			}
			val := b.Get([]byte(path))
			if val != nil {
				redirectURL = string(val)
			}
			return nil
		})

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if redirectURL != "" {
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}

		fallback.ServeHTTP(w, r)
	}
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if url, ok := pathsToUrls[path]; ok {
			http.Redirect(w, r, url, http.StatusFound)
			return
		}
		fallback.ServeHTTP(w, r)
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//   - path: /some-path
//     url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, db *bolt.DB, bucketName string, fallback http.Handler) (http.HandlerFunc, error) {
	parsedYaml, err := parseYaml(yml)
	if err != nil {
		return nil, err
	}

	err = storeInDB(parsedYaml, db, bucketName)
	if err != nil {
		return nil, err
	}

	return DBHandler(db, bucketName, fallback), nil

	// pathMap := buildMap(parsedYaml)
	//
	// return MapHandler(pathMap, fallback), nil
}

func JSONHandler(json []byte, db *bolt.DB, bucketName string, fallback http.Handler) (http.HandlerFunc, error) {
	parsedJson, err := parseJson(json)
	if err != nil {
		return nil, err
	}

	err = storeInDB(parsedJson, db, bucketName)
	if err != nil {
		return nil, err
	}

	return DBHandler(db, bucketName, fallback), nil

	// pathMap := buildMap(parsedJson)
	//
	// return MapHandler(pathMap, fallback), nil
}

func storeInDB(pairs []PathUrl, db *bolt.DB, bucketName string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		for _, pair := range pairs {
			err := bucket.Put([]byte(pair.Path), []byte(pair.URL))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func buildMap(pathUrls []PathUrl) map[string]string {
	pathsToUrls := make(map[string]string)
	for _, pu := range pathUrls {
		pathsToUrls[pu.Path] = pu.URL
	}
	return pathsToUrls
}

func parseYaml(data []byte) ([]PathUrl, error) {
	var pathUrls []PathUrl
	err := yaml.Unmarshal(data, &pathUrls)
	if err != nil {
		return nil, err
	}
	return pathUrls, nil
}

func parseJson(data []byte) ([]PathUrl, error) {
	var pathUrls []PathUrl
	err := json.Unmarshal(data, &pathUrls)
	if err != nil {
		return nil, err
	}
	return pathUrls, nil
}

type PathUrl struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url" json:"url"`
}
