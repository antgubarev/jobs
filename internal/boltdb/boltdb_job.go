package boltdb

import (
	"encoding/json"
	"fmt"

	"github.com/antgubarev/pet/internal/job"
	bolt "go.etcd.io/bbolt"
)

type JobStorage struct {
	db *bolt.DB
}

const JobBucketName string = "jobs"

func NewJobStorage(boltDB *bolt.DB) (*JobStorage, error) {
	if err := CreateBucketIfNotExists(boltDB, JobBucketName); err != nil {
		return nil, err
	}

	return &JobStorage{
		db: boltDB,
	}, nil
}

func (s *JobStorage) Store(job *job.Job) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := s.GetBucket(tx)
		if err != nil {
			return err
		}

		data, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("job store: marshal: %w", err)
		}

		if err := bucket.Put(s.GetJobKey(job.Name), data); err != nil {
			return fmt.Errorf("job store: bucket put: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("Store: %w", err)
	}

	return nil
}

func (s *JobStorage) GetByName(name string) (*job.Job, error) {
	var result job.Job

	if err := s.db.View(func(tx *bolt.Tx) error {
		bucket, err := s.GetBucket(tx)
		if err != nil {
			return err
		}
		data := bucket.Get(s.GetJobKey(name))
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("getbyname: unmarshal job: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("GetByName: %w", err)
	}

	return &result, nil
}

func (s *JobStorage) DeleteByName(name string) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := s.GetBucket(tx)
		if err != nil {
			return err
		}
		if err := bucket.Delete(s.GetJobKey(name)); err != nil {
			return fmt.Errorf("delete job: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("DeleteByName: %w", err)
	}

	return nil
}

func (s *JobStorage) GetAll() ([]job.Job, error) {
	var result []job.Job

	if err := s.db.View(func(tx *bolt.Tx) error {
		bucket, err := s.GetBucket(tx)
		if err != nil {
			return err
		}

		if err := bucket.ForEach(func(k, v []byte) error {
			var j job.Job
			if err := json.Unmarshal(v, &j); err != nil {
				return fmt.Errorf("getbyname: unmarshal job: %w", err)
			}
			result = append(result, j)

			return nil
		}); err != nil {
			return fmt.Errorf("search job in bucket: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("GetAll: %w", err)
	}

	return result, nil
}

func (s *JobStorage) GetBucket(tx *bolt.Tx) (*bolt.Bucket, error) {
	bucket := tx.Bucket([]byte(JobBucketName))
	if bucket == nil {
		return nil, fmt.Errorf("%w: %s", errBucketNotFound, JobBucketName)
	}

	return bucket, nil
}

func (s *JobStorage) GetJobKey(jobName string) []byte {
	return []byte(fmt.Sprintf("job:%s", jobName))
}
