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
	t.Helper()
	db = internal.NewTestBoltDB(t)
	store, err := boltdb.NewJobStorage(db)
	if err != nil {
		t.Errorf("new test job storage: %v", err)
	}

	return store, db
}

func TestBoltDbStorageStore(t *testing.T) {
	t.Parallel()
	store, testDB := newTestJobStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(testDB)

	testJob := &job.Job{
		Name:     "job",
		LockMode: job.HostLockMode,
	}

	err := store.Store(testJob)
	assert.NoError(t, err)

	if err = testDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(boltdb.JobBucketName))
		assert.NotNil(t, bucket)
		jBytes := bucket.Get(store.GetJobKey(testJob.Name))
		var j job.Job
		err = json.Unmarshal(jBytes, &j)
		assert.NoError(t, err)
		changelog, _ := diff.Diff(testJob, j)
		assert.True(t, len(changelog) == 0, "jobs are not equal: %s", internal.DiffToString(&changelog))

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	jobUpdated := &job.Job{
		Name:     "job",
		LockMode: job.ClusterLockMode,
	}

	err = store.Store(jobUpdated)
	assert.NoError(t, err)

	err = testDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(boltdb.JobBucketName))
		assert.NotNil(t, bucket)
		jBytes := bucket.Get(store.GetJobKey(jobUpdated.Name))
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
	t.Parallel()
	store, testDB := newTestJobStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(testDB)

	jobs := make([]job.Job, 4)
	jobs[0] = job.Job{
		Name:     "job1",
		LockMode: job.HostLockMode,
	}
	jobs[1] = job.Job{
		Name:     "job2",
		LockMode: job.HostLockMode,
	}

	if err := storeJobs(jobs, store, testDB); err != nil {
		t.Errorf("store fixtures %v", err)
	}

	jb, err := store.GetByName("job1")
	assert.NoError(t, err)
	assert.Equal(t, jobs[0].Name, jb.Name, 0)
	assert.Equal(t, jobs[0].LockMode, jb.LockMode, 0)
}

func TestBoltDbStorageDeleteByName(t *testing.T) {
	t.Parallel()
	testJobStorage, testDB := newTestJobStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(testDB)

	jobs := make([]job.Job, 2)
	jobs[0] = job.Job{
		Name:     "job1",
		LockMode: job.HostLockMode,
	}
	jobs[1] = job.Job{
		Name:     "job2",
		LockMode: job.HostLockMode,
	}

	if err := storeJobs(jobs, testJobStorage, testDB); err != nil {
		t.Errorf("store fixtures %v", err)
	}

	err := testJobStorage.DeleteByName("job1")
	assert.NoError(t, err)

	if err := testDB.View(func(tx *bolt.Tx) error {
		bucket, err := testJobStorage.GetBucket(tx)
		if err != nil {
			return fmt.Errorf("view job in bucket: %w", err)
		}

		if err := bucket.ForEach(func(k, v []byte) error {
			var j job.Job
			if err := json.Unmarshal(v, &j); err != nil {
				t.Errorf("unmarshal job: %v", err)
			}
			assert.Equal(t, "job2", j.Name)

			return nil
		}); err != nil {
			return fmt.Errorf("search job in bucket: %w", err)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func storeJobs(jobs []job.Job, store *boltdb.JobStorage, db *bolt.DB) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := store.GetBucket(tx)
		if err != nil {
			return fmt.Errorf("get bucket: %w", err)
		}
		for _, j := range jobs {
			data, err := json.Marshal(j)
			if err != nil {
				return fmt.Errorf("job marshal: %w", err)
			}
			if err := bucket.Put(store.GetJobKey(j.Name), data); err != nil {
				return fmt.Errorf("job put to bucket: %w", err)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("store fixtures %w", err)
	}

	return nil
}
