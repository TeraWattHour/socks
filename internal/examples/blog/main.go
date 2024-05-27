package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
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
		"env":         "test",
		"Metas":       []string{"author: TeraWattHour", time.Now().String()},
		"currentDate": time.Now(),
	}); err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
		if err != nil || page < 1 {
			page = 1
		}

		if err := s.Execute(w, "index.html", map[string]interface{}{
			"currentDate": time.Now(),
			"page":        int(page),
		}); err != nil {
			w.WriteHeader(500)
			log.Println(err)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
