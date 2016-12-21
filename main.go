package main

import (
	"encoding/binary"
	"log"
	"net/http"

	"encoding/json"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/boltdb/bolt"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/todos", GetTodo),
		rest.Put("/todos", PutTodo),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

type todo struct {
	Did   bool   `json:"did"`
	Title string `json:"title"`
}

func GetTodo(w rest.ResponseWriter, r *rest.Request) {
	db, err := bolt.Open("my.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	todos, err := query(db, "todos")
	if err != nil {
		w.WriteJson(map[string]string{"msg": err.Error()})
		return
	}

	w.WriteJson(todos)
}

func query(db *bolt.DB, bucket string) ([]todo, error) {
	var t todo
	todos := []todo{}

	if err := db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if err := json.Unmarshal(v, &t); err == nil {
				todos = append(todos, t)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return todos, nil
}

func PutTodo(w rest.ResponseWriter, r *rest.Request) {
	var t todo
	err := r.DecodeJsonPayload(&t)
	if err != nil {
		w.WriteJson(map[string]string{"msg": err.Error()})
		return
	}

	db, err := bolt.Open("my.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	insert(db, "todos", t)
}

func insert(db *bolt.DB, bucket string, v todo) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		// Create a bucket.
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		id, _ := b.NextSequence()

		bv, err := json.Marshal(v)
		if err != nil {
			return err
		}

		// Set the value "bar" for the key "foo".
		if err := b.Put([]byte(itob(int(id))), bv); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
