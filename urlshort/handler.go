package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	for path := range pathsToUrls {
		fmt.Printf("Registering path: %s\n", path)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if redirect, ok := pathsToUrls[r.URL.Path]; ok {
			fmt.Printf("Redirecting %s → %s\n", r.URL.Path, redirect)
			http.Redirect(w, r, redirect, http.StatusMovedPermanently)
		} else {
			fallback.ServeHTTP(w, r)
		}
	})
}

type Redirect struct {
	Path string
	Url  string
}

func toPathMap(paths []Redirect) map[string]string {
	return lo.FromEntries(lo.Map(paths, func(path Redirect, _ int) lo.Entry[string, string] {
		return lo.Entry[string, string]{Key: path.Path, Value: path.Url}
	}))
}

func YamlHandler(yml []byte, fallback http.Handler) (http.Handler, error) {
	redirects := []Redirect{}

	if err := yaml.Unmarshal(yml, &redirects); err != nil {
		return nil, err
	}
	pathsToUrls := toPathMap(redirects)

	return MapHandler(pathsToUrls, fallback), nil
}

func JsonHandler(yml []byte, fallback http.Handler) (http.Handler, error) {
	redirects := []Redirect{}

	if err := json.Unmarshal(yml, &redirects); err != nil {
		return nil, err
	}
	pathsToUrls := toPathMap(redirects)

	return MapHandler(pathsToUrls, fallback), nil
}
