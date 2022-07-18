package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	bolt "go.etcd.io/bbolt"
	"gopkg.in/yaml.v3"
)

type RouteHandler interface {
	MapHandler(pathsToUrls map[string]string) RouteHandler
	ServeHTTP(http.ResponseWriter, *http.Request)
	Close()
}
type memoryEngine http.HandlerFunc

func NewMemoryEngine(handler http.Handler) RouteHandler {
	return memoryEngine(handler.ServeHTTP)
}

func (handler memoryEngine) MapHandler(pathsToUrls map[string]string) RouteHandler {
	for path := range pathsToUrls {
		fmt.Printf("Registering path: %s\n", path)
	}

	return memoryEngine(func(w http.ResponseWriter, r *http.Request) {
		if redirect, ok := pathsToUrls[r.URL.Path]; ok {
			fmt.Printf("Redirecting %s → %s\n", r.URL.Path, redirect)
			http.Redirect(w, r, redirect, http.StatusMovedPermanently)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}

func (handler memoryEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler(w, r)
}

func (memoryEngine) Close() {
}

type boltEngine struct {
	db      *bolt.DB
	file    *os.File
	handler http.HandlerFunc
}

func CreateBoltEngine(handler http.Handler) (RouteHandler, error) {
	file, err := ioutil.TempFile(".", "bolt-*.db")
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(file.Name(), 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("Redirects"))
		return err
	})
	return boltEngine{db, file, handler.ServeHTTP}, nil
}

func (engine boltEngine) MapHandler(pathsToUrls map[string]string) RouteHandler {
	err := engine.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Redirects"))

		for path, redirect := range pathsToUrls {
			fmt.Printf("Registering path: %s\n", path)
			err := b.Put([]byte(path), []byte(redirect))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return boltEngine{engine.db, engine.file, func(w http.ResponseWriter, r *http.Request) {
		engine.db.View(func(tx *bolt.Tx) error {

			redirect := tx.Bucket([]byte("Redirects")).Get([]byte(r.URL.Path))

			if redirect != nil {
				fmt.Printf("Redirecting %s → %s\n", r.URL.Path, string(redirect))
				http.Redirect(w, r, string(redirect), http.StatusMovedPermanently)
			} else {
				engine.handler.ServeHTTP(w, r)
			}

			return nil
		})
	}}
}

func (engine boltEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	engine.handler(w, r)
}

func (engine boltEngine) Close() {
	fmt.Println("Closing bolt engine")
	engine.db.Close()
	os.Remove(engine.file.Name())
}

type redirect struct {
	Path string
	Url  string
}

func YamlHandler(yml []byte, handler RouteHandler) (RouteHandler, error) {
	redirects := []redirect{}

	if err := yaml.Unmarshal(yml, &redirects); err != nil {
		return nil, err
	}

	return handler.MapHandler(toPathMap(redirects)), nil
}

func JsonHandler(yml []byte, handler RouteHandler) (RouteHandler, error) {
	redirects := []redirect{}

	if err := json.Unmarshal(yml, &redirects); err != nil {
		return nil, err
	}

	return handler.MapHandler(toPathMap(redirects)), nil
}

func toPathMap(redirects []redirect) (pathsToUrls map[string]string) {
	pathsToUrls = make(map[string]string)
	for _, redirect := range redirects {
		pathsToUrls[redirect.Path] = redirect.Url
	}
	return
}
