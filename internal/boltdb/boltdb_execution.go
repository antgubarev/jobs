package boltdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/antgubarev/pet/internal/job"
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

type ExecutionStorage struct {
	db *bolt.DB
}

func NewExecutionStorage(db *bolt.DB) (*ExecutionStorage, error) {
	if err := CreateBucketIfNotExists(db, JobBucketName); err != nil {
		return nil, err
	}
	return &ExecutionStorage{db: db}, nil
}

func (bes *ExecutionStorage) Store(execution *job.Execution) error {
	return bes.db.Update(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return err
		}

		data, err := json.Marshal(execution)
		if err != nil {
			return fmt.Errorf("execution store: marshal: %v", err)
		}

		if err := bucket.Put(bes.GetExecutionKey(execution), data); err != nil {
			return fmt.Errorf("execution store: bucket put: %v", err)
		}

		return nil
	})
}

func (bes *ExecutionStorage) GetById(id uuid.UUID) (job.Execution, error) {
	var result job.Execution
	return result, bes.db.View(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return err
		}
		bucket.ForEach(func(k, v []byte) error {
			var e job.Execution
			if err := json.Unmarshal(v, &e); err != nil {
				return fmt.Errorf("execution getbyname: unmarshal job: %v", err)
			}
			if e.Id == id {
				result = e
				return nil
			}
			return nil
		})
		return nil
	})
}

func (bes *ExecutionStorage) GetByJobName(jobName string) ([]job.Execution, error) {
	var result []job.Execution
	return result, bes.db.View(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return err
		}
		c := bucket.Cursor()
		prefix := bes.GetExecutionNameKeyPrefix(jobName)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var e job.Execution
			if err := json.Unmarshal(v, &e); err != nil {
				return fmt.Errorf("execution getbyname: unmarshal job: %v", err)
			}
			result = append(result, e)
		}
		return nil
	})
}

func (bes *ExecutionStorage) DeleteByJobName(jobName string) error {
	return bes.db.Update(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return err
		}
		c := bucket.Cursor()
		prefix := bes.GetExecutionNameKeyPrefix(jobName)
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := c.Delete(); err != nil {
				return fmt.Errorf("execution remove: %v", err)
			}
		}
		return nil
	})
}

func (bes *ExecutionStorage) Delete(execution *job.Execution) error {
	return bes.db.Update(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return err
		}
		c := bucket.Cursor()
		prefix := bes.GetExecutionNameKeyPrefix(execution.Job)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			var e job.Execution
			if err := json.Unmarshal(v, &e); err != nil {
				return fmt.Errorf("execution delete: unmarshal job: %v", err)
			}

			if *e.Host == *execution.Host && *e.Pid == *execution.Pid {
				if err := c.Delete(); err != nil {
					return fmt.Errorf("execution remove: %v", err)
				}
				return nil
			}
		}
		return errors.New("execution delete: execution not found")
	})
}

func (bes *ExecutionStorage) GetBucket(tx *bolt.Tx) (*bolt.Bucket, error) {
	bucket := tx.Bucket([]byte(JobBucketName))
	if bucket == nil {
		return nil, fmt.Errorf("bucket %s doesn't exist", JobBucketName)
	}
	return bucket, nil
}

func (bes *ExecutionStorage) GetExecutionKey(execution *job.Execution) []byte {
	host := ""
	if execution.Host != nil {
		host = *execution.Host
	}
	pid := 0
	if execution.Pid != nil {
		pid = *execution.Pid
	}
	return []byte(fmt.Sprintf("execution:%s:%s:%d", execution.Job, host, pid))
}

func (bes *ExecutionStorage) GetExecutionNameKeyPrefix(name string) []byte {
	return []byte(fmt.Sprintf("execution:%s:", name))
}
