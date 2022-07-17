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
				Usage:     "A yaml file that contains the mapping of paths to urls",
				TakesFile: true,
				Aliases:   []string{"y"},
			},
		},
		Action: func(c *cli.Context) error {

			mux := defaultMux()

			// Build the MapHandler using the mux as the fallback
			pathsToUrls := map[string]string{
				"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
				"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
			}
			handler := MapHandler(pathsToUrls, mux)

			yamlPaths := c.StringSlice("yaml")

			for _, yamlPath := range yamlPaths {
				yaml, err := os.ReadFile(yamlPath)
				if err != nil {
					return err
				}

				handler, err = YamlHandler([]byte(yaml), handler)
				if err != nil {
					return err
				}
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
