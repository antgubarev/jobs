package internal

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/antgubarev/jobs/internal/boltdb"
	"github.com/gin-gonic/gin"
	"github.com/r3labs/diff/v2"
	bolt "go.etcd.io/bbolt"
)

func DiffToString(changeLog *diff.Changelog) string {
	var result string
	for _, change := range *changeLog {
		result += fmt.Sprintf("'%s' expected: %v, Actual: %v \n", strings.Join(change.Path, "."), change.From, change.To)
	}

	return result
}

func CheckDifferErrors(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("check diff error: %v", err)
	}
}

func NewTestBoltDB(t *testing.T) *bolt.DB {
	t.Helper()
	dir, err := ioutil.TempDir("", "boltdb_test")
	if err != nil {
		t.Errorf("creating temp dir: %v", err)
	}

	file, err := ioutil.TempFile(dir, "*.dat")
	if err != nil {
		t.Errorf("creating temp file: %v", err)
	}
	db, err := boltdb.NewBoltDB(file.Name())
	if err != nil {
		t.Errorf("creating bolt file %v", err)
	}

	return db
}

func NewTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	return gin.New()
}
