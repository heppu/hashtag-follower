package db

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

type Client struct {
	db     *bolt.DB
	bucket []byte
}

func NewClient(file, bucket string) (c *Client, err error) {
	var db *bolt.DB
	if db, err = bolt.Open(file, 0600, &bolt.Options{Timeout: 30 * time.Second}); err != nil {
		log.Fatal(err)
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}); err != nil {
		return
	}

	c = &Client{
		db:     db,
		bucket: []byte(bucket),
	}
	return
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) GetTags(chatID int64) (map[string]struct{}, error) {
	hashtags := make(map[string]struct{})
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.bucket)
		id := make([]byte, 8)
		binary.LittleEndian.PutUint64(id, uint64(chatID))
		return gob.NewDecoder(bytes.NewBuffer(b.Get(id))).Decode(&hashtags)
	})
	return hashtags, err
}

func (c *Client) AddTag(chatID int64, hashtag string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.bucket)
		hashtags := make(map[string]struct{})
		id := make([]byte, 8)
		binary.LittleEndian.PutUint64(id, uint64(chatID))
		tagsBuf := b.Get(id)

		if len(tagsBuf) > 0 {
			if err := gob.NewDecoder(bytes.NewBuffer(tagsBuf)).Decode(&hashtags); err != nil {
				return err
			}
		}

		hashtags[hashtag] = struct{}{}
		buf := &bytes.Buffer{}
		if err := gob.NewEncoder(buf).Encode(hashtags); err != nil {
			return err
		}
		return b.Put(id, buf.Bytes())
	})
}

func (c *Client) DeleteTag(chatID int64, hashtag string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.bucket)
		hashtags := make(map[string]struct{})
		id := make([]byte, 8)
		binary.LittleEndian.PutUint64(id, uint64(chatID))
		if err := gob.NewDecoder(bytes.NewBuffer(b.Get(id))).Decode(&hashtags); err != nil {
			return err
		}
		if _, ok := hashtags[hashtag]; !ok {
			return nil
		}
		delete(hashtags, hashtag)
		buf := &bytes.Buffer{}
		if err := gob.NewEncoder(buf).Encode(hashtags); err != nil {
			return err
		}
		return b.Put(id, buf.Bytes())
	})
}
