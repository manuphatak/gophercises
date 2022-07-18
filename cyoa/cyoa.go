package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/mux"
)

type Arc struct {
	Title   string    `json:"title"`
	Story   []string  `json:"story"`
	Options []Options `json:"options"`
}
type Options struct {
	Text string `json:"text"`
	Arc  string `json:"arc"`
}

func main() {
	const tpl = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8" />
		<meta http-equiv="X-UA-Compatible" content="IE=edge" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<title>{{ .Title }}</title>
	</head>
	<body>
		<h1>{{ .Title }}</h1>
		{{ range .Story }}<p>{{ . }}</p>{{ end }}

		<ol>
			{{ range .Options }}<li><a href='/gopher/{{ .Arc }}'>{{ .Text }}</a></li>{{ end }}
		</ol>

	</body>
</html>`
	t, err := template.New("foo").Parse(tpl)
	check(err)

	jsonFile, err := os.Open("gopher.json")
	check(err)

	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	check(err)

	story := map[string]Arc{}

	err = json.Unmarshal(jsonData, &story)
	check(err)

	r := mux.NewRouter()

	r.HandleFunc("/gopher", http.RedirectHandler("/gopher/intro", http.StatusFound).ServeHTTP)

	r.HandleFunc("/gopher/{arc}", func(w http.ResponseWriter, r *http.Request) {
		arcKey, ok := mux.Vars(r)["arc"]
		if !ok {
			return
		}
		arc := story[arcKey]

		err = t.Execute(w, arc)
		check(err)
	})

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", r)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
