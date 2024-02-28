package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	openai "github.com/sashabaranov/go-openai"
)

type MoviesList struct {
	List1 []string
	List2 []string
}

type Input struct {
	Name   string `json:"name"`
	Column int    `json:"column"`
}

type Movie struct {
	Name       string `json:"name"`
	Likelihood string `json:"likelihood"`
}

type Response struct {
	CommonMovies []Movie `json:"common_movies"`
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

	apiRouter.HandleFunc("/get-films", processMovies).Methods("POST")

	http.Handle("/", r)

	log.Print("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func processMovies(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	if apiKey == "" {
		log.Println("No OpenAI API key found. Set the OPENAI_API_KEY environment variable.")
	}

	client := openai.NewClient(apiKey)

	movies := MoviesList{
		List1: make([]string, 0),
		List2: make([]string, 0),
	}

	// Process each input
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
			if column == 1 {
				movies.List1 = append(movies.List1, value)
			} else {
				movies.List2 = append(movies.List2, value)
			}
		}
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(`I will provide two lists of movies that two different people like. 
					You must respond with a list of three different movies that both people would like in JSON format 
					with the name of the movie and a percentage likelihood of both people liking them:
                    List 1: %q
                    List 2: %q

                    common_movies: [
                       {
                         "name": <movie_name>,
                         "likelihood": <percentage>
                        },
                        {
                         "name": <movie_name>,
                         "likelihood": <percentage>
                        },
                        {
                         "name": <movie_name>,
                         "likelihood": <percentage>
                         }
                    ]`, movies.List1, movies.List2),
				},
			},
		},
	)

	if err != nil {
		fmt.Fprintf(w, "<h1 class='text-red-500'>Something went wrong: %v</h1>", err)
	}

	var response Response
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &response)
	if err != nil {
		fmt.Fprintf(w, "Error parsing chat response: %v", err)
	}

	const htmlTemplate = `
	<div class="container mt-3">
		<div class="row">
			<div class="col">
				<h3>Common Movies</h3>
				<ul class="list-group">
					{{range .}}
						<li class="list-group-item">
							<strong>{{.Name}}</strong> - Likelihood: {{.Likelihood}}
						</li>
					{{end}}
				</ul>
			</div>
		</div>
	</div>`

	tmpl, err := template.New("common_movies").Parse(htmlTemplate)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	var builder strings.Builder
	err = tmpl.Execute(&builder, response.CommonMovies)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	fmt.Fprint(w, builder.String())
}
