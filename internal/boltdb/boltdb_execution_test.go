package boltdb_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/antgubarev/pet/internal"
	"github.com/antgubarev/pet/internal/boltdb"
	"github.com/antgubarev/pet/internal/job"
	"github.com/google/uuid"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func newTestExecutionStorage(t *testing.T) (store *boltdb.ExecutionStorage, db *bolt.DB) {
	db = internal.NewTestBoltDb(t)
	store, err := boltdb.NewExecutionStorage(db)
	if err != nil {
		t.Errorf("new test job storage: %v", err)
	}
	return store, db
}

func TestBoltDbExecutionStorageStore(t *testing.T) {
	store, db := newTestExecutionStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(db)

	execution := &job.Execution{
		Job:        "job",
		Command:    internal.NewPointerOfString("command"),
		Pid:        internal.NewPointerOfInt(1),
		Host:       internal.NewPointerOfString("host"),
		StartedAt:  time.Now(),
		FinishedAt: internal.NewPointerOfTime(time.Now()),
		Status:     job.StatusRunning,
		Msg:        internal.NewPointerOfString("msg"),
	}

	err := store.Store(execution)
	assert.NoError(t, err)

	if err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(boltdb.JobBucketName))
		assert.NotNil(t, bucket)
		jBytes := bucket.Get([]byte(store.GetExecutionKey(execution)))
		var e job.Execution
		err = json.Unmarshal(jBytes, &e)
		assert.NoError(t, err)
		changelog, _ := diff.Diff(execution, e)
		assert.True(t, len(changelog) == 0, "jobs are not equal: %s", internal.DiffToString(&changelog))
		return nil
	}); err != nil {
		t.Errorf("view bucket: %v", err)
	}

	executionUpdated := &job.Execution{
		Job:        "job",
		Command:    internal.NewPointerOfString("commandUpdated"),
		Pid:        internal.NewPointerOfInt(1),
		Host:       internal.NewPointerOfString("host"),
		StartedAt:  time.Now().Add(time.Hour),
		FinishedAt: internal.NewPointerOfTime(time.Now().Add(time.Hour)),
		Status:     job.StatusFailed,
		Msg:        internal.NewPointerOfString("msgUpdated"),
	}

	err = store.Store(executionUpdated)
	assert.NoError(t, err)

	if err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(boltdb.JobBucketName))
		assert.NotNil(t, bucket)
		jBytes := bucket.Get([]byte(store.GetExecutionKey(execution)))
		var e job.Execution
		err = json.Unmarshal(jBytes, &e)
		assert.NoError(t, err)
		changelog, _ := diff.Diff(execution, e)
		assert.True(t, len(changelog) == 0, "jobs are not equal: %s", internal.DiffToString(&changelog))
		return nil
	}); err != nil {
		t.Errorf("view bucket: %v", err)
	}
}

func TestBoltDbExecutionStorageGetByJobName(t *testing.T) {
	store, db := newTestExecutionStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(db)

	executions := make([]job.Execution, 4)
	executions[0] = job.Execution{
		Job:        "job",
		Command:    internal.NewPointerOfString("command"),
		Pid:        internal.NewPointerOfInt(1),
		Host:       internal.NewPointerOfString("host1"),
		StartedAt:  time.Now(),
		FinishedAt: internal.NewPointerOfTime(time.Now()),
		Status:     job.StatusRunning,
		Msg:        internal.NewPointerOfString("msg"),
	}
	executions[1] = job.Execution{
		Job:        "job",
		Command:    internal.NewPointerOfString("command"),
		Pid:        internal.NewPointerOfInt(1),
		Host:       internal.NewPointerOfString("host2"),
		StartedAt:  time.Now(),
		FinishedAt: internal.NewPointerOfTime(time.Now()),
		Status:     job.StatusRunning,
		Msg:        internal.NewPointerOfString("msg"),
	}
	executions[2] = job.Execution{
		Job:        "job",
		Command:    internal.NewPointerOfString("command"),
		Pid:        internal.NewPointerOfInt(1),
		Host:       internal.NewPointerOfString("host3"),
		StartedAt:  time.Now(),
		FinishedAt: internal.NewPointerOfTime(time.Now()),
		Status:     job.StatusRunning,
		Msg:        internal.NewPointerOfString("msg"),
	}
	executions[3] = job.Execution{
		Job:        "job2",
		Command:    internal.NewPointerOfString("command"),
		Pid:        internal.NewPointerOfInt(1),
		Host:       internal.NewPointerOfString("host"),
		StartedAt:  time.Now(),
		FinishedAt: internal.NewPointerOfTime(time.Now()),
		Status:     job.StatusRunning,
		Msg:        internal.NewPointerOfString("msg"),
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := store.GetBucket(tx)
		if err != nil {
			t.Errorf("get bucket: %v", err)
		}
		for _, e := range executions {
			data, err := json.Marshal(e)
			if err != nil {
				t.Errorf("execution marshal: %v", err)
			}
			if err := bucket.Put(store.GetExecutionKey(&e), data); err != nil {
				t.Errorf("job put to bucket: %v", err)
			}
		}
		return nil
	}); err != nil {
		t.Errorf("store fixtures %v", err)
	}

	executionsByName, err := store.GetByJobName("job")
	assert.NoError(t, err)
	assert.Len(t, executionsByName, 3)
	for _, e := range executionsByName {
		diff1, err := diff.Diff(e, executions[0])
		internal.CheckDifferErrors(t, err)

		diff2, err := diff.Diff(e, executions[1])
		internal.CheckDifferErrors(t, err)

		diff3, err := diff.Diff(e, executions[0])
		internal.CheckDifferErrors(t, err)

		if len(diff1) == 0 && len(diff2) == 0 && len(diff3) == 0 {
			t.Errorf("unexpected job")
		}
	}
}

func TestBoltDbExecutionGetById(t *testing.T) {
	store, db := newTestExecutionStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(db)

	id := uuid.New()
	original := job.Execution{
		Id:        id,
		Job:       "job",
		StartedAt: time.Now(),
		Status:    job.StatusRunning,
	}
	if err := store.Store(&original); err != nil {
		t.Error(err)
	}

	execution, err := store.GetById(id)
	assert.NoError(t, err)
	assert.Equal(t, original.Id, execution.Id)
}

func TestBoltDbExecutionDelete(t *testing.T) {
	store, db := newTestExecutionStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(db)

	executions := make([]job.Execution, 4)
	executions[0] = job.Execution{
		Job:        "job",
		Command:    internal.NewPointerOfString("command"),
		Pid:        internal.NewPointerOfInt(1),
		Host:       internal.NewPointerOfString("host1"),
		StartedAt:  time.Now(),
		FinishedAt: internal.NewPointerOfTime(time.Now()),
		Status:     job.StatusRunning,
		Msg:        internal.NewPointerOfString("msg"),
	}
	executions[1] = job.Execution{
		Job:        "job",
		Command:    internal.NewPointerOfString("command"),
		Pid:        internal.NewPointerOfInt(1),
		Host:       internal.NewPointerOfString("host2"),
		StartedAt:  time.Now(),
		FinishedAt: internal.NewPointerOfTime(time.Now()),
		Status:     job.StatusRunning,
		Msg:        internal.NewPointerOfString("msg"),
	}

	err := store.Store(&executions[0])
	assert.NoError(t, err)

	err = store.Store(&executions[1])
	assert.NoError(t, err)

	err = store.Delete(&job.Execution{
		Job:     "job",
		Command: nil,
		Pid:     internal.NewPointerOfInt(1),
		Host:    internal.NewPointerOfString("host1"),
		Status:  job.StatusRunning,
	})
	assert.NoError(t, err)

	items, err := store.GetByJobName("job")
	if err != nil {
		t.Errorf("get executions items %v", err)
	}
	assert.Equal(t, 1, len(items))
}
