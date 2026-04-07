package main

import (
	"fmt"
	"net/http"
)

const POST_LEN int = 10

func main() {

	mux := http.NewServeMux()

	db := createDB("localhost:6379", "", 0, 2)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		short_path := r.URL.Path[1:] // gets path and removes leading slash
		original_url, err := db.getValue(short_path)
		if err != nil {
			fmt.Fprintln(w, "URL not found")
		}
		fmt.Fprintln(w, "Short path: " + short_path)
		fmt.Fprintln(w, "Original URL: " + original_url)
		if err == nil{
			http.Redirect(w, r, original_url, http.StatusFound)
		}

	})

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		reader := r.Body
		message := make([]byte, POST_LEN)
		reader.Read(message)
		fmt.Fprintf(w, "Message: %s\n", string(message)) // DEBUG
		msg_str := string(message)
		short_url, err := db.addURL(msg_str)
		if err != nil {
			fmt.Fprintln(w, "URL couldn't be added")
			panic(err)
		}

		url_bytes := []byte(short_url)
		w.Write(url_bytes)
		fmt.Fprintln(w, short_url)

	})

	mux.HandleFunc("DELETE /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "DELETE /")
	})

	fmt.Println("Started server...")


	http.ListenAndServe(":8080", mux)
}


func checkShortURL(short_url string, db DBStruct) (string, error) {

	return db.getValue(short_url)
}

