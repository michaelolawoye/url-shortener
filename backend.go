package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

const POST_LEN int = 10

// json format structs
type CreateURLRequest struct {
	Url string `json:"url"`
}
type CreateURLResponse struct {
	Status bool `json:"status"`
	Desc string `json:"desc"`
	Url string `json:"url"`
}
type deleteKeyRequest struct {
	Url string
	Short bool `json:"short"`
}
type deleteKeyResponse struct {
	Status bool `json:"status"`
	Desc string `json:"desc"`
}
type deleteAllKeysResponse struct {
	Status bool `json:"status"`
	Desc string `json:"desc"`
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

		req_decoder := json.NewDecoder(r.Body)
		resp_encoder := json.NewEncoder(w)

		if err := req_decoder.Decode(&reqdata); err != nil {
			s := fmt.Sprintf("Error occurred while decoding json request\nError: %s\n", err)
			respdata = CreateURLResponse{Status: false, Desc: s}
			if err = resp_encoder.Encode(respdata); err != nil {
				fmt.Printf("postRoot: Couldn't send response to client\nError: %s\n", err)
			}
			return 
		}

		fmt.Fprintf(w, "struct: %s.\n", reqdata) // DEBUG
		
		
		short_url, err := db.addURL(reqdata.Url)
		if err != nil {
			respdata = CreateURLResponse{Status:false, Desc:"Failed to add url to database", Url:""}
			if err = resp_encoder.Encode(respdata); err != nil {
				fmt.Printf("postRoot: Server couldn't send response\nError: %s\n", err)
			}
			return
		}

		respdata = CreateURLResponse{Status:true, Desc:"Successful", Url:short_url}

		if err = resp_encoder.Encode(respdata); err != nil {
			fmt.Printf("postRoot: Server couldn't send response\nError: %s\n", err)
			return
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
	
	w.Header().Set("Content-Type", "application/json")
	var reqdata deleteKeyRequest
	var respdata deleteKeyResponse

	req_decoder := json.NewDecoder(r.Body)
	resp_encoder := json.NewEncoder(w)

	if err := req_decoder.Decode(&reqdata); err != nil {
		s := fmt.Sprintf("Error occurred while decoding client request\nError: %s\n", err)
		respdata = deleteKeyResponse{Status:false, Desc: s}
		if err = resp_encoder.Encode(&respdata); err != nil {
			fmt.Printf("deleteSingleKey: Couldn't send response to client\nError: %s\n", err)
		}
		return
	}

	fmt.Fprintf(w, "URL to be deleted: %s.\n", string(reqdata.Url)) // DEBUG


	deleted_val, err := db.removeURL(reqdata.Url, false)
	if err != nil {
		s := fmt.Sprintf("Failed to remove URL\nError: %s\n", err)
		respdata = deleteKeyResponse{Status: false, Desc: s}
		if err = resp_encoder.Encode(&respdata); err != nil {
			fmt.Printf("deleteSingleKey: Couldn't send response to client\nError: %s\n", err)
		}
		return
	}	

	fmt.Fprintln(w, "Deleted value: " + deleted_val) // DEBUG

	respdata = deleteKeyResponse{Status: true, Desc: "Successful"}

	if err = json.NewEncoder(w).Encode(&respdata); err != nil {
		fmt.Printf("deleteSingleKey: Couldn't send response to client\nError: %s\n", err)
	}

}

func (db *DBStruct) deleteAllKeys(w http.ResponseWriter, r *http.Request) {

	var respdata deleteAllKeysResponse
	resp_encoder := json.NewEncoder(w)

	keys, err := db.client.Keys(db.ctx, "*").Result()
	if err != nil {
		s := fmt.Sprintf("Couldn't get keys from database\nError: %s\n", err)
		respdata = deleteAllKeysResponse{Status: false, Desc: s}
		if err = resp_encoder.Encode(&respdata); err != nil {
			fmt.Printf("deleteAllKeys: Couldn't send response to client\nError: %s\n", err)
		}
		return
	}

	for _, key := range keys {
		_, err = db.deleteEntry(key, false)
		if err != nil {
			s := fmt.Sprintf("Couldn't delete key: %s", key)
			respdata = deleteAllKeysResponse{Status: false, Desc: s}
			if err = resp_encoder.Encode(&respdata); err != nil {
				fmt.Printf("deleteAllKeys: Couldn't send response to client\nError: %s\n", err)
				return
			}
		}
	}
	respdata = deleteAllKeysResponse{Status: true, Desc: "Finished"}
	if err := resp_encoder.Encode(&respdata); err != nil {
		fmt.Printf("deleteAllKeys: Couldn't send response to client\nError: %s\n", err)
	}

}


func checkShortURL(short_url string, db DBStruct) (string, error) {

	return db.getValue(short_url, false)
}
