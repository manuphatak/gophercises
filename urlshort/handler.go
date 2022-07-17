package main

import (
	"net/http"

	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if redirect, ok := pathsToUrls[r.URL.Path]; ok {
			http.Redirect(w, r, redirect, http.StatusMovedPermanently)
		} else {
			fallback.ServeHTTP(w, r)
		}
	}
}

type PathMap struct {
	Path string
	Url  string
}

// YamlHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YamlHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	paths := []PathMap{}

	if err := yaml.Unmarshal(yml, &paths); err != nil {
		return nil, err
	}
	pathsToUrls := lo.FromEntries(lo.Map(paths, func(path PathMap, _ int) lo.Entry[string, string] {
		return lo.Entry[string, string]{Key: path.Path, Value: path.Url}
	}))

	return MapHandler(pathsToUrls, fallback), nil
}
