package main

import (
	"fmt"
	"net/http"
)

func main() {

	mux := http.NewServeMux()

	db := createDB("localhost:6379", "", 0, 2)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		short_path := r.URL.Path
		short_path = short_path[1:]
		original_url, err := db.getValue(short_path)
		if err != nil {
			fmt.Fprintln(w, "URL not found")
		}
		fmt.Fprintln(w, "Short path: " + short_path)
		fmt.Fprintln(w, "Original URL: " + original_url)
		
	})
	fmt.Println("Started server...")

	http.ListenAndServe(":8080", mux)
}


func checkShortURL(short_url string, db DBStruct) (string, error) {

	return db.getValue(short_url)
}

