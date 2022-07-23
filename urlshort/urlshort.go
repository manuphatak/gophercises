package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	var handler RouteHandler

	app := &cli.App{
		Name:                   "urlshort",
		Usage:                  "A url shortener",
		Compiled:               time.Now(),
		HideHelpCommand:        true,
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		ArgsUsage:              " ",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:      "yaml",
				Value:     cli.NewStringSlice("paths.yaml"),
				Usage:     "A YAML file that contains the mapping of paths to urls",
				TakesFile: true,
				Aliases:   []string{"y"},
			},
			&cli.StringSliceFlag{
				Name:      "json",
				Value:     cli.NewStringSlice(),
				Usage:     "A JSON file that contains the mapping of paths to urls",
				TakesFile: true,
				Aliases:   []string{"j"},
			},

			&cli.BoolFlag{
				Name:    "bolt",
				Value:   false,
				Usage:   "Use a bolt-db backend to store and serve the redirects",
				Aliases: []string{"b"},
			},
		},
		Action: func(c *cli.Context) error {
			var err error
			mux := defaultMux()

			if c.Bool("bolt") {
				handler, err = CreateBoltEngine(mux)
				if err != nil {
					return err
				}
			} else {
				handler = NewMemoryEngine(mux)
			}
			defer handler.Close()

			handler = loadDefaultRedirects(handler)

			yamlPaths := c.StringSlice("yaml")
			handler, err = loadYamlRedirects(yamlPaths, handler)
			if err != nil {
				return err
			}

			jsonPaths := c.StringSlice("json")
			handler, err = loadJsonRedirects(jsonPaths, handler)
			if err != nil {
				return err
			}

			fmt.Println("Starting the server on :8080")
			http.ListenAndServe(":8080", handler)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	})
	return mux
}

func loadDefaultRedirects(handler RouteHandler) RouteHandler {
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	// Build the MapHandler using the mux as the fallback
	return handler.MapHandler(pathsToUrls)
}

func loadYamlRedirects(yamlPaths []string, handler RouteHandler) (RouteHandler, error) {
	for _, yamlPath := range yamlPaths {
		yaml, err := os.ReadFile(yamlPath)
		if err != nil {
			return nil, err
		}

		handler, err = YamlHandler([]byte(yaml), handler)
		if err != nil {
			return nil, err
		}
	}
	return handler, nil
}

func loadJsonRedirects(jsonPaths []string, handler RouteHandler) (RouteHandler, error) {
	for _, jsonPath := range jsonPaths {
		json, err := os.ReadFile(jsonPath)
		if err != nil {
			return nil, err
		}

		handler, err = JsonHandler([]byte(json), handler)
		if err != nil {
			return nil, err
		}
	}
	return handler, nil
}
