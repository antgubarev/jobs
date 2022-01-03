package boltdb_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/antgubarev/pet/internal"
	"github.com/antgubarev/pet/internal/boltdb"
	"github.com/antgubarev/pet/internal/job"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func newTestJobStorage(t *testing.T) (store *boltdb.JobStorage, db *bolt.DB) {
	db = internal.NewTestBoltDb(t)
	store, err := boltdb.NewJobStorage(db)
	if err != nil {
		t.Errorf("new test job storage: %v", err)
	}
	return store, db
}

func TestBoltDbStorageStore(t *testing.T) {
	store, db := newTestJobStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(db)

	jb := &job.Job{
		Name:     "job",
		LockMode: job.HostLockMode,
	}

	err := store.Store(jb)
	assert.NoError(t, err)

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(boltdb.JobBucketName))
		assert.NotNil(t, bucket)
		jBytes := bucket.Get([]byte(store.GetJobKey(jb.Name)))
		var j job.Job
		err = json.Unmarshal(jBytes, &j)
		assert.NoError(t, err)
		changelog, _ := diff.Diff(jb, j)
		assert.True(t, len(changelog) == 0, "jobs are not equal: %s", internal.DiffToString(&changelog))
		return nil
	})

	jobUpdated := &job.Job{
		Name:     "job",
		LockMode: job.ClusterLockMode,
	}

	err = store.Store(jobUpdated)
	assert.NoError(t, err)

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(boltdb.JobBucketName))
		assert.NotNil(t, bucket)
		jBytes := bucket.Get([]byte(store.GetJobKey(jobUpdated.Name)))
		var j job.Job
		err = json.Unmarshal(jBytes, &j)
		assert.NoError(t, err)
		changelog, _ := diff.Diff(*jobUpdated, j)
		assert.True(t, len(changelog) == 0, "jobs are not equal: %s", internal.DiffToString(&changelog))
		return nil
	})
	assert.NoError(t, err)
}

func TestBoltDbStorageStoreGetByName(t *testing.T) {
	store, db := newTestJobStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(db)

	jobs := make([]job.Job, 4)
	jobs[0] = job.Job{
		Name:     "job1",
		LockMode: job.HostLockMode,
	}
	jobs[1] = job.Job{
		Name:     "job2",
		LockMode: job.HostLockMode,
	}

	if err := storeJobs(jobs, store, db); err != nil {
		t.Errorf("store fixtures %v", err)
	}

	jb, err := store.GetByName("job1")
	assert.NoError(t, err)
	assert.Equal(t, jobs[0].Name, jb.Name, 0)
	assert.Equal(t, jobs[0].LockMode, jb.LockMode, 0)
}

func TestBoltDbStorageDeleteByName(t *testing.T) {
	store, db := newTestJobStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(db)

	jobs := make([]job.Job, 2)
	jobs[0] = job.Job{
		Name:     "job1",
		LockMode: job.HostLockMode,
	}
	jobs[1] = job.Job{
		Name:     "job2",
		LockMode: job.HostLockMode,
	}

	if err := storeJobs(jobs, store, db); err != nil {
		t.Errorf("store fixtures %v", err)
	}

	err := store.DeleteByName("job1")
	assert.NoError(t, err)

	db.View(func(tx *bolt.Tx) error {
		bucket, err := store.GetBucket(tx)
		if err != nil {
			return err
		}

		bucket.ForEach(func(k, v []byte) error {
			var j job.Job
			if err := json.Unmarshal(v, &j); err != nil {
				t.Errorf("unmarshal job: %v", err)
			}
			assert.Equal(t, "job2", j.Name)
			return nil
		})

		return nil
	})
}

func storeJobs(jobs []job.Job, store *boltdb.JobStorage, db *bolt.DB) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := store.GetBucket(tx)
		if err != nil {
			return fmt.Errorf("get bucket: %v", err)
		}
		for _, j := range jobs {
			data, err := json.Marshal(j)
			if err != nil {
				return fmt.Errorf("job marshal: %v", err)
			}
			if err := bucket.Put(store.GetJobKey(j.Name), data); err != nil {
				return fmt.Errorf("job put to bucket: %v", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("store fixtures %v", err)
	}
	return nil
}
