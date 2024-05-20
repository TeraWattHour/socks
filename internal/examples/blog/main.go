package main

import (
	"log"
	"net/http"
	"time"

	"github.com/terawatthour/socks"
)

type Post struct {
	Title    string
	Comments []string
}

func main() {
	s := socks.NewSocks(&socks.Options{
		Sanitizer: func(s string) string {
			return s
		},
	})

	if err := s.LoadTemplates("internal/examples/blog/templates/*.html", "internal/examples/blog/templates"); err != nil {
		panic(err)
	}

	if err := s.Compile(map[string]any{
		"Title": "TeraWattHour's blog",
		"Menus": []string{"Home", "About", "Contact"},
		"Posts": []Post{
			{"Hello Wordl", []string{"Nice post!", "I like it!"}},
			{"Goodbye World", []string{"Sad to see you go.", "Good luck!"}},
		},
		"Metas":       []string{"author: TeraWattHour", time.Now().String()},
		"currentDate": time.Now(),
	}); err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := s.Execute(w, "index.html", map[string]interface{}{
			"currentDate": time.Now(),
		}); err != nil {
			w.WriteHeader(500)
			log.Println(err)
		}
	})

	http.ListenAndServe(":8080", nil)
}
