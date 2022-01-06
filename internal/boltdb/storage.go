package boltdb

import (
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

var (
	errExecutionNotFound = errors.New("execution ot found")
	errBucketNotFound    = errors.New("bucket not found")
)

const BoltdbFileAccess = 0666

func NewBoltDB(path string) (*bolt.DB, error) {
	boltDB, err := bolt.Open(path, BoltdbFileAccess, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to boltdb %w", err)
	}

	if err = boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(JobBucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("create bucket %w", err)
	}

	return boltDB, nil
}

func CreateBucketIfNotExists(db *bolt.DB, bucketName string) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("boltDb Tx: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("CreateBucketIfNotExists: %w", err)
	}

	return nil
}
