package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

const POST_LEN int = 10

type CreateURLRequest struct {
	Url string `json:"url"`
}
type CreateURLResponse struct {
	Status bool `json:"status"`
	Desc string `json:"desc"`
	Url string `json:"url"`
}
type deleteKeyStruct struct {
	Url string
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
		// DEBUG ----------
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

		w.Header().Set("Content-Type", "application/json")

		var respdata CreateURLResponse
		var reqdata CreateURLRequest

		if err := json.NewDecoder(r.Body).Decode(&reqdata); err != nil { 
			fmt.Fprintf(w, "Error occurred while decoding json request\nError: %s\n", err)
			return 
		}

		fmt.Fprintf(w, "struct: %s.\n", reqdata) // DEBUG
		
		
		short_url, err := db.addURL(reqdata.Url)
		if err != nil {
			respdata = CreateURLResponse{Status:false, Desc:"Failed to add url to database", Url:""}
			if err = json.NewEncoder(w).Encode(respdata); err != nil {
				fmt.Printf("Server couldn't send response\nError: %s\n", err)
			}
			return
		}

		respdata = CreateURLResponse{Status:true, Desc:"Successful", Url:short_url}

		if err = json.NewEncoder(w).Encode(respdata); err != nil {
			fmt.Printf("Server couldn't send response\nError: %s\n", err)
		}

		// DEBUG ----------
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
	
	var reqdata deleteKeyStruct
	if err := json.NewDecoder(r.Body).Decode(&reqdata); err != nil {
		fmt.Fprintln(w, "Data must be in correct json format")
		return
	}

	fmt.Fprintf(w, "URL to be deleted: %s.\n", string(reqdata.Url)) // DEBUG


	deleted_val, err := db.removeURL(reqdata.Url, false)
	if err != nil {
		fmt.Fprintln(w, "Couldn't remove url from database")
		return
	}	

	fmt.Fprintln(w, "Deleted value: " + deleted_val) // DEBUG

	respdata := deleteKeyStruct{Url: deleted_val}

	if err = json.NewEncoder(w).Encode(&respdata); err != nil {
		fmt.Fprintln(w, "Server couldn't send response to delete key request")
		return
	}

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

