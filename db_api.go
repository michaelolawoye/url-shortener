package main

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type DBStruct struct {
	client *redis.Client;
	ctx context.Context;
}

func createDB(addr string, password  string, db int, protocol int) DBStruct {
	database := DBStruct{
		redis.NewClient(&redis.Options{
			Addr: addr,
			Password: password,
			DB: db,
			Protocol: protocol,
		}), context.Background()}

	return database
}

func (db DBStruct) addURL(url string) (string, error) { // adds url to database as a value, with key being the generated short url
	short_url, err := db.generateShortURL(url)
	if err != nil {
		return "", err
	}

	err = db.setEntry(short_url, url)
	if err != nil {
		return "", err
	}

	return short_url, nil
}

func (db DBStruct) setEntry(key string, value string) error {

	err := db.client.Set(db.ctx, key, value, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (db DBStruct) getValue(key string) (string, error) {

	val, err := db.client.Get(db.ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil

}

func (db DBStruct) closeDB() {
	defer db.client.Close()
}

func (db DBStruct) generateShortURL(long_url string) (string, error) {


	var sum byte = 0
	byts := []byte(long_url)
	for _, i := range byts {
		sum += i
	}

	var seed []byte = []byte{2, 3, 5, 7, 11, 13, 17, 19}

	for i, v := range seed {
		new_val := (v*sum % 50) + 72
		if new_val >= 91 && new_val <=96 {
			new_val += 7
		}
		seed[i] = new_val
	}

	short_url := string(seed)

	return short_url, nil
}