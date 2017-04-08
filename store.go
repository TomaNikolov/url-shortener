package main

import (
	"bytes"

	"github.com/boltdb/bolt"
)

// Panic ...
var Panic = func(v interface{}) {
	panic(v)
}

// Store ...
type Store interface {
	Set(key string, value string) error //error if something went wrong
	Get(key string) string              // empty value if not found
	Len() int                           // should retutn the number of all records
	GetAll() map[string]string          // should retutn the number of all records
	Close() error                       // release the store or ignore
}

var (
	tableURLs = []byte("urls")
)

// DB ...
type DB struct {
	db *bolt.DB
}

var _ Store = &DB{}

func openDatabase(stumb string) *bolt.DB {
	db, err := bolt.Open(stumb, 0600, nil)
	if err != nil {
		Panic(err)
	}

	// create the buckets here
	var tables = [...][]byte{
		tableURLs,
	}

	db.Update(func(tx *bolt.Tx) (err error) {
		for _, table := range tables {
			_, err = tx.CreateBucketIfNotExists(table)
			if err != nil {
				Panic(err)
			}
		}

		return
	})

	return db
}

// NewDB returns a new DB instance, its connection is opened.
// DB implements the Store.
func NewDB(stumb string) *DB {
	return &DB{
		db: openDatabase(stumb),
	}
}

// Set ...
func (d *DB) Set(key string, value string) error {
	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(tableURLs)
		// Generate ID for the url
		// Note: we could use that instead of a random string key
		// but we want to simulate a real-world url shortener
		// so we skip that.
		// id, _ := b.NextSequence()
		return b.Put([]byte(key), []byte(value))
	})

	return nil
}

// Get ...
func (d *DB) Get(key string) (value string) {
	keyb := []byte(key)
	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(tableURLs)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if bytes.Equal(keyb, k) {
				value = string(v)
				break
			}
		}

		return nil
	})

	return
}

// Len returns all the "shorted" urls length
func (d *DB) Len() (num int) {
	d.db.View(func(tx *bolt.Tx) error {

		// Assume bucket exists and has keys
		b := tx.Bucket(tableURLs)

		b.ForEach(func([]byte, []byte) error {
			num++
			return nil
		})
		return nil
	})
	return
}

// GetAll ...
func (d *DB) GetAll() (m map[string]string) {
	d.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(tableURLs)
		m = make(map[string]string)
		b.ForEach(func(k []byte, v []byte) error {
			m[string(k[:])] = string(v[:])
			return nil
		})
		return nil
	})
	return
}

// Close the data(base) connection
func (d *DB) Close() error {
	return d.db.Close()
}
