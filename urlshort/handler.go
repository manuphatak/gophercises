package main

import (
	"net/http"

	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

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
