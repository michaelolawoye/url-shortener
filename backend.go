package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"os"
)

const POST_LEN int = 10

// json format structs
type GetURLResponse struct {
	Status bool `json:"status"`
	Desc string `json:"desc"`
	Url string `json:"url"`
	Entry string `json:"entry"`
}
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
	ShortUrl bool `json:"shorturl"`
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

	addr := os.Getenv("REDIS_HOST")
	if addr == "" {
		addr = "localhost"
	} 
	fmt.Println("Address recieved: " + addr)
	db := createDB(addr + ":6379", "", 0, 2)
	defer db.closeDB()

	mux.HandleFunc("GET /", db.getRoot) // leave just root url (/) for full database, add short url for specific original url

	mux.HandleFunc("POST /create-url", db.postRoot)

	mux.HandleFunc("DELETE /single-key", db.deleteSingleKey)

	mux.HandleFunc("DELETE /all-keys", db.deleteAllKeys)


	fmt.Println("Started server...")
	http.ListenAndServe(":8080", mux)

}


func (db *DBStruct) getRoot(w http.ResponseWriter, r *http.Request) {

	var respdata GetURLResponse

	resp_encoder := json.NewEncoder(w)

	short_url := r.URL.Path[1:]

	if len(short_url) == 0 {
		if err := getURLDatabase(&respdata, resp_encoder, db); err != nil {
			s := fmt.Sprintf("Couldn't get database values\nError: %s\n", err)
			respdata = GetURLResponse{Status: false, Desc: s}
			if err = resp_encoder.Encode(&respdata); err != nil {
				responseFailError("getRoot", err)
			}
		}
	} else {
		if err := getSingleKey(&respdata, resp_encoder, db, short_url); err != nil {
			s := fmt.Sprintf("Couldn't get key from database\n Error: %s\n", err)
			respdata = GetURLResponse{Status: false, Desc: s}
			if err = resp_encoder.Encode(&respdata); err != nil {
				responseFailError("getRoot", err)
			}
		}
	}

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

func getURLDatabase(respdata *GetURLResponse, resp_encoder *json.Encoder, db *DBStruct) error {

	key_entries, err := db.client.Keys(db.ctx, "*").Result()
	if err != nil {
		return err
	}


	for e := range key_entries {
		res, err := db.client.Get(db.ctx, key_entries[e]).Result()
		if err != nil {
			return err
		}
		s := fmt.Sprintf("%s:%s", key_entries[e], res)
		respdata = &GetURLResponse{Status: true, Desc: "Successful", Url: "", Entry: s}
		if err = resp_encoder.Encode(respdata); err != nil {
			return err
		}
	}
	respdata = &GetURLResponse{Status: true, Desc: "Finished", Url: "", Entry: ""}
	if err := resp_encoder.Encode(respdata); err != nil {
		return err
	}

	return nil
}

func getSingleKey(respdata *GetURLResponse, resp_encoder *json.Encoder, db *DBStruct, short_url string) error {

	original_url, err := db.getEntry(short_url, false); 
	if err != nil {
		return err
	}

	respdata = &GetURLResponse{Status: true, Desc: "Successful", Url: original_url, Entry: ""}

	if err = resp_encoder.Encode(respdata); err != nil {
		return err
	}

	return nil
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
			if err = resp_encoder.Encode(&respdata); err != nil {
				responseFailError("postRoot", err)
			}
			return 
		}

		fmt.Fprintf(w, "struct: %s.\n", reqdata) // DEBUG
		
		
		short_url, err := db.addURL(reqdata.Url)
		if err != nil {
			respdata = CreateURLResponse{Status:false, Desc:"Failed to add url to database", Url:""}
			if err = resp_encoder.Encode(&respdata); err != nil {
				responseFailError("postRoot", err)
			}
			return
		}

		respdata = CreateURLResponse{Status:true, Desc:"Successful", Url:short_url}

		if err = resp_encoder.Encode(&respdata); err != nil {
			responseFailError("postRoot", err)
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
			responseFailError("deleteSingleKey", err)
		}
		return
	}

	fmt.Fprintf(w, "URL to be deleted: %s.\n", string(reqdata.Url)) // DEBUG


	deleted_val, err := db.removeURL(reqdata.Url, false)
	if err != nil {
		s := fmt.Sprintf("Failed to remove URL\nError: %s\n", err)
		respdata = deleteKeyResponse{Status: false, Desc: s}
		if err = resp_encoder.Encode(&respdata); err != nil {
			responseFailError("deleteSingleKey", err)
		}
		return
	}	

	fmt.Fprintln(w, "Deleted value: " + deleted_val) // DEBUG

	respdata = deleteKeyResponse{Status: true, Desc: "Successful"}

	if err = json.NewEncoder(w).Encode(&respdata); err != nil {
		responseFailError("deleteSingleKey", err)
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
			responseFailError("deleteAllKeys", err)
		}
		return
	}

	for _, key := range keys {
		_, err = db.deleteEntry(key, false)
		if err != nil {
			s := fmt.Sprintf("Couldn't delete key: %s", key)
			respdata = deleteAllKeysResponse{Status: false, Desc: s}
			if err = resp_encoder.Encode(&respdata); err != nil {
				responseFailError("deleteAllKeys", err)
				return
			}
		}
	}
	respdata = deleteAllKeysResponse{Status: true, Desc: "Finished"}
	if err := resp_encoder.Encode(&respdata); err != nil {
		responseFailError("deleteAllKeys", err)
	}

}

func checkShortURL(short_url string, db *DBStruct) (string, error) {

	return db.getEntry(short_url, false)
}
func responseFailError(funcname string, err error) {
	fmt.Printf("%s: Couldn't send response to client\nError: %s\n", funcname, err)
}
