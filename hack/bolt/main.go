package main

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/glower/bakku-app/pkg/storage"
)

func main() {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		panic(err)
	}
	// defer

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			return err
		}
		return b.Put([]byte("answer"), []byte("42"))
	})
	if err != nil {
		panic(err)
	}

	var value []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		value = b.Get([]byte("answer"))
		return nil
	})
	fmt.Printf("The final answer is: %s\n", value)
	db.Close()

	s := storage.New("my.db")
	v, err := s.Get("answer", "MyBucket")
	if err != nil {
		panic(err)
	}
	fmt.Printf("The final answer is: %s\n", v)
}
