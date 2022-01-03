package boltdb

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func NewBoltDb(path string) (*bolt.DB, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to boltdb %v", err)
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(JobBucketName))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("create bucket %v", err)
	}

	return db, nil
}

func CreateBucketIfNotExists(db *bolt.DB, bucketName string) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("create bucket %v", err)
	}
	return nil
}
