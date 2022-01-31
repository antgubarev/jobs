package boltdb

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/antgubarev/jobs/internal/job"
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
	// TODO: Id!!!
	if err := bes.db.Update(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return err
		}

		data, err := json.Marshal(execution)
		if err != nil {
			return fmt.Errorf("execution store: marshal: %w", err)
		}

		if err := bucket.Put(bes.GetExecutionKey(execution), data); err != nil {
			return fmt.Errorf("execution store: bucket put: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("Store execution: %w", err)
	}

	return nil
}

func (bes *ExecutionStorage) GetByID(executionID uuid.UUID) (*job.Execution, error) {
	var result job.Execution

	err := bes.db.View(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return fmt.Errorf("get bucket: %w", err)
		}
		if err := bucket.ForEach(func(k, v []byte) error {
			var e job.Execution
			if err := json.Unmarshal(v, &e); err != nil {
				return fmt.Errorf("execution getbyname: unmarshal job: %w", err)
			}
			if e.ID == executionID {
				result = e

				return nil
			}

			return nil
		}); err != nil {
			return fmt.Errorf("search execution in bucket: %w", err)
		}

		return nil
	})
	if err != nil {
		return &result, fmt.Errorf("GetByID: %w", err)
	}

	return &result, nil
}

func (bes *ExecutionStorage) GetByJobName(jobName string) ([]job.Execution, error) {
	var result []job.Execution

	if err := bes.db.View(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return err
		}
		c := bucket.Cursor()
		prefix := bes.GetExecutionNameKeyPrefix(jobName)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var e job.Execution
			if err := json.Unmarshal(v, &e); err != nil {
				return fmt.Errorf("execution getbyname: unmarshal job: %w", err)
			}
			result = append(result, e)
		}

		return nil
	}); err != nil {
		return result, fmt.Errorf("GetJobByName: %w", err)
	}

	return result, nil
}

func (bes *ExecutionStorage) DeleteByJobName(jobName string) error {
	if err := bes.db.Update(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return err
		}
		c := bucket.Cursor()
		prefix := bes.GetExecutionNameKeyPrefix(jobName)
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := c.Delete(); err != nil {
				return fmt.Errorf("execution remove: %w", err)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("DeleteJobByName: %w", err)
	}

	return nil
}

func (bes *ExecutionStorage) Delete(executionID uuid.UUID) error {
	if err := bes.db.Update(func(tx *bolt.Tx) error {
		bucket, err := bes.GetBucket(tx)
		if err != nil {
			return err
		}
		if err := bucket.ForEach(func(key, value []byte) error {
			var e job.Execution
			if err := json.Unmarshal(value, &e); err != nil {
				return fmt.Errorf("unmarshal execution: %w", err)
			}
			if e.ID.String() == executionID.String() {
				if err := bucket.Delete(key); err != nil {
					return fmt.Errorf("remove execution from bucket: %w", err)
				}

				return nil
			}

			return nil
		}); err != nil {
			return fmt.Errorf("search execution in bucket: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("Delete execution: %w", err)
	}

	return nil
}

func (bes *ExecutionStorage) GetBucket(tx *bolt.Tx) (*bolt.Bucket, error) {
	bucket := tx.Bucket([]byte(JobBucketName))
	if bucket == nil {
		return nil, fmt.Errorf("GetBucket %s: %w", JobBucketName, errBucketNotFound)
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
