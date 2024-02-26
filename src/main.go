package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type MoviesList struct {
	List1 []string
	List2 []string
}

type Input struct {
	Name   string `json:"name"`
	Column int    `json:"column"`
}

func main() {
	r := mux.NewRouter()
	tmpl := template.Must(template.ParseFiles("src/templates/index.html"))

	apiRouter := r.PathPrefix("/api").Subrouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}
	})

	apiRouter.HandleFunc("/get-films", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Process each input
		var inputs []Input
		for key, values := range r.Form {
			for _, value := range values {
				// Extract column number from the input name
				// Assuming input names are in the format "inputX_columnY" where X is the input number and Y is the column number
				// Split the key by underscore to get the input and column parts
				parts := strings.Split(key, "_")
				if len(parts) != 2 {
					// Skip this input if the format is invalid
					continue
				}
				// Extract the column number from the second part of the key
				column, err := strconv.Atoi(parts[1][len(parts[1])-1:])
				if err != nil {
					log.Fatal(err)
				}
				inputs = append(inputs, Input{Name: value, Column: column})
			}
		}

		fmt.Fprintf(w, "Your lists:\n")
		for _, input := range inputs {
			fmt.Fprintf(w, "Name: %s, Column: %d\n", input.Name, input.Column)
		}
	}).Methods("POST")

	http.Handle("/", r)

	log.Print("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
