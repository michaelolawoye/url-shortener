package main

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type DBStruct struct {
	client *redis.Client;
	reverse_client *redis.Client;
	ctx context.Context;
}

func createDB(addr string, password  string, db int, protocol int) DBStruct {
	database := DBStruct{
		redis.NewClient(&redis.Options{
			Addr: addr,
			Password: password,
			DB: db,
			Protocol: protocol,
		}),
		redis.NewClient(&redis.Options{
			Addr: addr,
			Password: password,
			DB: db+1,
			Protocol: protocol,
		}),
		context.Background()}

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

func (db DBStruct) removeURL(short_url string, reverse bool) (string, error) {

	return db.deleteEntry(short_url, reverse)
}

func (db DBStruct) setEntry(key string, value string) error {

	err := db.client.Set(db.ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	err = db.reverse_client.Set(db.ctx, value, key, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (db DBStruct) deleteEntry(key string, reverse bool) (string, error) {

	val, err := db.client.Get(db.ctx, key).Result()
	if err != nil {
		return "get", err
	}

	err = db.client.Del(db.ctx, key).Err()
	if err != nil {
		return "del", err
	}
	err = db.reverse_client.Del(db.ctx, val).Err()
	if err != nil {
		return "reverse-del", err
	}

	return val, nil
}

func (db DBStruct) getEntry(key string, reverse bool) (string, error) {

	var val string
	var err error
	if reverse { // retrieve from the reverse db
		val, err = db.reverse_client.Get(db.ctx, key).Result()
		if err != nil {
			return "", err
		}
	} else { // retrieve from the normal db
		val, err = db.client.Get(db.ctx, key).Result()
		if err != nil {
			return "", err
		}
}
	return val, nil

}

func (db DBStruct) closeDB() {
	defer db.client.Close()
	defer db.reverse_client.Close()
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