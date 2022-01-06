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
	t.Helper()

	db = internal.NewTestBoltDB(t)
	store, err := boltdb.NewExecutionStorage(db)
	if err != nil {
		t.Errorf("new test job storage: %v", err)
	}

	return store, db
}

func TestBoltDbExecutionStorageStore(t *testing.T) {
	t.Parallel()
	testExecutionStorage, testDB := newTestExecutionStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(testDB)

	execution := job.NewRunningExecution("job")
	execution.SetID(uuid.Nil)
	execution.SetCommand("command")
	execution.SetPid(1)
	execution.SetHost("host")

	err := testExecutionStorage.Store(execution)
	assert.NoError(t, err)

	viewStorage(t, execution, testExecutionStorage, testDB)

	executionUpdated := job.NewRunningExecution("job")
	executionUpdated.SetID(uuid.Nil)
	executionUpdated.SetCommand("commandUpdated")
	executionUpdated.SetPid(1)
	executionUpdated.SetHost("host")
	executionUpdated.Finish(job.StatusFailed, time.Now().Add(time.Hour), "msgUpdated")

	err = testExecutionStorage.Store(executionUpdated)
	assert.NoError(t, err)

	viewStorage(t, execution, testExecutionStorage, testDB)
}

func viewStorage(t *testing.T, execution *job.Execution, store *boltdb.ExecutionStorage, db *bolt.DB) {
	t.Helper()

	if err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(boltdb.JobBucketName))
		assert.NotNil(t, bucket)
		jBytes := bucket.Get(store.GetExecutionKey(execution))
		var e job.Execution
		err := json.Unmarshal(jBytes, &e)
		assert.NoError(t, err)
		changelog, _ := diff.Diff(execution, e)
		assert.True(t, len(changelog) == 0, "jobs are not equal: %s", internal.DiffToString(&changelog))

		return nil
	}); err != nil {
		t.Errorf("view bucket: %v", err)
	}
}

func TestBoltDbExecutionStorageGetByJobName(t *testing.T) {
	t.Parallel()
	store, testDB := newTestExecutionStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(testDB)

	executions := make([]job.Execution, 4)
	executions[0] = *job.NewRunningExecution("job")
	executions[0].SetID(uuid.Nil)
	executions[0].SetCommand("command")
	executions[0].SetPid(1)
	executions[0].SetHost("host1")

	executions[1] = *job.NewRunningExecution("job")
	executions[1].SetID(uuid.Nil)
	executions[1].SetCommand("command")
	executions[1].SetPid(1)
	executions[1].SetHost("host2")

	executions[2] = *job.NewRunningExecution("job")
	executions[2].SetID(uuid.Nil)
	executions[2].SetCommand("command")
	executions[2].SetPid(1)
	executions[2].SetHost("host3")

	executions[3] = *job.NewRunningExecution("job2")
	executions[3].SetID(uuid.Nil)
	executions[3].SetCommand("command")
	executions[3].SetPid(1)
	executions[3].SetHost("host1")

	if err := testDB.Update(func(tx *bolt.Tx) error {
		bucket, err := store.GetBucket(tx)
		if err != nil {
			t.Errorf("get bucket: %v", err)
		}
		for i := range executions {
			data, err := json.Marshal(executions[i])
			if err != nil {
				t.Errorf("execution marshal: %v", err)
			}
			if err := bucket.Put(store.GetExecutionKey(&executions[i]), data); err != nil {
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
	for _, exec := range executionsByName {
		diff1, err := diff.Diff(exec, executions[0])
		internal.CheckDifferErrors(t, err)

		diff2, err := diff.Diff(exec, executions[1])
		internal.CheckDifferErrors(t, err)

		diff3, err := diff.Diff(exec, executions[0])
		internal.CheckDifferErrors(t, err)

		if len(diff1) == 0 && len(diff2) == 0 && len(diff3) == 0 {
			t.Errorf("unexpected job")
		}
	}
}

func TestBoltDbExecutionGetById(t *testing.T) {
	t.Parallel()
	store, db := newTestExecutionStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(db)

	executionID := uuid.New()
	original := job.NewRunningExecution("job")
	original.SetCommand("command")
	original.SetPid(1)
	original.SetHost("host1")

	if err := store.Store(original); err != nil {
		t.Error(err)
	}

	execution, err := store.GetByID(executionID)
	assert.NoError(t, err)
	assert.Equal(t, original.ID, execution.ID)
}

func TestBoltDbExecutionDelete(t *testing.T) {
	t.Parallel()
	store, db := newTestExecutionStorage(t)
	defer func(db *bolt.DB) {
		db.Close()
		os.Remove(db.Path())
	}(db)

	executions := make([]*job.Execution, 4)
	executions[0] = job.NewRunningExecution("job")
	executions[0].SetCommand("command")
	executions[0].SetPid(1)
	executions[0].SetHost("host1")

	executions[1] = job.NewRunningExecution("job")
	executions[1].SetCommand("command")
	executions[1].SetPid(1)
	executions[1].SetHost("host2")

	err := store.Store(executions[0])
	assert.NoError(t, err)

	err = store.Store(executions[1])
	assert.NoError(t, err)

	executionForDelete := job.NewRunningExecution("job")
	executionForDelete.SetID(uuid.Nil)
	executionForDelete.SetPid(1)
	executionForDelete.SetHost("host1")

	err = store.Delete(executionForDelete)
	assert.NoError(t, err)

	items, err := store.GetByJobName("job")
	if err != nil {
		t.Errorf("get executions items %v", err)
	}
	assert.Equal(t, 1, len(items))
}
