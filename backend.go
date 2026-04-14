package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

const POST_LEN int = 10

type CreateURLStruct struct {

	url string
}

func main() {

	mux := http.NewServeMux()

	db := createDB("localhost:6379", "", 0, 2)
	defer db.closeDB()

	mux.HandleFunc("GET /", db.getRoot)

	mux.HandleFunc("POST /", db.postRoot)

	mux.HandleFunc("DELETE /single-key", db.deleteSingleKey)

	mux.HandleFunc("DELETE /all-keys", db.deleteAllKeys)


	fmt.Println("Started server...")
	http.ListenAndServe(":8080", mux)

}


func (db *DBStruct) getRoot(w http.ResponseWriter, r *http.Request) {
		short_path := r.URL.Path[1:] // gets path and removes leading slash

		// DEBUG ----------
		fmt.Fprintln(w, "Short URL: " + short_path)
		fmt.Fprintln(w, "Current database:")
		key_entries, err := db.client.Keys(db.ctx, "*").Result()
		if err != nil {
			fmt.Fprintln(w, "Couldn't print database entries")
			panic(err)
		}
		for e := range key_entries {
			res, err := db.client.Get(db.ctx,key_entries[e]).Result()	
			if err != nil {
				fmt.Fprintf(w, "Couldn't get value from key: %s", key_entries[e])
			}
			fmt.Fprintf(w, "%s: %s\n", key_entries[e], res)

		}

		// DEBUG ----------

}

func (db *DBStruct) postRoot(w http.ResponseWriter, r *http.Request) {

		var reqdata CreateURLStruct
		if err := json.NewDecoder(r.Body).Decode(&reqdata); err != nil {
			fmt.Fprintln(w, "Data must be in correct json format")
			return 
		}

		fmt.Fprintf(w, "URL: %s\n", reqdata.url) // DEBUG
		
		
		short_url, err := db.addURL(reqdata.url)
		if err != nil {
			fmt.Fprintln(w, "URL couldn't be added")
			return
		}

		var respdata CreateURLStruct

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(respdata); err != nil {
			fmt.Fprintln(w, "Server couldn't send response")
		}

		// DEBUG ----------
		fmt.Fprintln(w, "Short URL: " + short_url)
		fmt.Fprintln(w, "Current database:")
		key_entries, err := db.client.Keys(db.ctx, "*").Result()
		if err != nil {
			fmt.Fprintln(w, "Couldn't get keys from database")
			return
		}
		for e := range key_entries {
			res, err := db.client.Get(db.ctx,key_entries[e]).Result()	
			if err != nil {
				fmt.Fprintf(w, "Couldn't get value from key: %s", key_entries[e])
			}
			fmt.Fprintf(w, "%s: %s\n", key_entries[e], res)

		}
		// DEBUG ----------

}

func (db *DBStruct) deleteSingleKey(w http.ResponseWriter, r *http.Request) {

		reader := r.Body

		message := make([]byte, POST_LEN)
		reader.Read(message)

		fmt.Fprintf(w, "URL to be deleted: %s.\n", string(message)) // DEBUG

		msg_str := string(message)

		deleted_val, err := db.removeURL(msg_str, false)
		if err != nil {
			fmt.Fprintln(w, "Couldn't remove url from database")
			return
		}	

		fmt.Fprintln(w, "Deleted value: " + deleted_val) // DEBUG

}

func (db *DBStruct) deleteAllKeys(w http.ResponseWriter, r *http.Request) {

		keys, err := db.client.Keys(db.ctx, "*").Result()
		if err != nil {
			fmt.Fprintln(w, "Couldn't get keys")
			return
		}

		for _, key := range keys {
			_, err = db.deleteEntry(key, false)
			if err != nil {
				fmt.Fprintln(w, "Couldn't delete key: " + key)
			}
		}

}


func checkShortURL(short_url string, db DBStruct) (string, error) {

	return db.getValue(short_url, false)
}

