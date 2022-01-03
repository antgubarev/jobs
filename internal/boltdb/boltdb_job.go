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

func NewJobStorage(db *bolt.DB) (*JobStorage, error) {
	if err := CreateBucketIfNotExists(db, JobBucketName); err != nil {
		return nil, err
	}

	return &JobStorage{
		db: db,
	}, nil
}

func (s *JobStorage) Store(job *job.Job) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := s.GetBucket(tx)
		if err != nil {
			return err
		}

		data, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("job store: marshal: %v", err)
		}

		if err := bucket.Put(s.GetJobKey(job.Name), data); err != nil {
			return fmt.Errorf("job store: bucket put: %v", err)
		}

		return nil
	})
}

func (s *JobStorage) GetByName(name string) (*job.Job, error) {
	var result job.Job
	return &result, s.db.View(func(tx *bolt.Tx) error {
		bucket, err := s.GetBucket(tx)
		if err != nil {
			return err
		}
		data := bucket.Get(s.GetJobKey(name))
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("getbyname: unmarshal job: %v", err)
		}
		return nil
	})
}

func (s *JobStorage) DeleteByName(name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := s.GetBucket(tx)
		if err != nil {
			return err
		}
		if err := bucket.Delete(s.GetJobKey(name)); err != nil {
			return fmt.Errorf("delete job: %v", err)
		}
		return nil
	})
}

func (s *JobStorage) GetAll() ([]job.Job, error) {
	var result []job.Job
	return result, s.db.View(func(tx *bolt.Tx) error {
		bucket, err := s.GetBucket(tx)
		if err != nil {
			return err
		}

		bucket.ForEach(func(k, v []byte) error {
			var j job.Job
			if err := json.Unmarshal(v, &j); err != nil {
				return fmt.Errorf("getbyname: unmarshal job: %v", err)
			}
			result = append(result, j)
			return nil
		})

		return nil
	})
}

func (s *JobStorage) GetBucket(tx *bolt.Tx) (*bolt.Bucket, error) {
	bucket := tx.Bucket([]byte(JobBucketName))
	if bucket == nil {
		return nil, fmt.Errorf("bucket %s doesn't exist", JobBucketName)
	}
	return bucket, nil
}

func (s *JobStorage) GetJobKey(jobName string) []byte {
	return []byte(fmt.Sprintf("job:%s", jobName))
}
