package executor_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/antgubarev/pet/internal/executor"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStartAndFinish(t *testing.T) {
	restClient := new(restapi.MockClient)
	restClient.On("JobStart", mock.MatchedBy(func(in *restapi.JobStartIn) bool {
		return in.Job == "job" && *in.Command == "sh ../../tests/fixtures/echo.sh 3 1"
	})).Return(uuid.New(), nil)

	outFile := createTestOutFile(t, "out_")
	defer func() {
		outFile.Close()
		os.Remove(outFile.Name())
	}()
	exectr := executor.NewExecutor(restClient, executor.WithOutFile(outFile))

	code, err := exectr.StartAndWatch(context.Background(), "job", []string{"sh", "../../tests/fixtures/echo.sh", "3", "1"})
	assert.Equal(t, 0, code)
	assert.NoError(t, err)

	realOutput := readOutFile(t, outFile)
	assert.Equal(t, "step 1\nstep 2\nfinish\n", string(realOutput))
}

func createTestOutFile(t *testing.T, prefix string) *os.File {
	dir, err := ioutil.TempDir("", "executor_test")
	if err != nil {
		t.Errorf("creating temp dir: %v", err)
	}

	file, err := ioutil.TempFile(dir, fmt.Sprintf("%s*.txt", prefix))
	if err != nil {
		t.Errorf("creating temp file: %v", err)
	}
	return file
}

func readOutFile(t *testing.T, f *os.File) string {
	output, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("read out file %v", err)
	}
	return string(output)
}

func TestStartAndSendSigint(t *testing.T) {
	restClient := new(restapi.MockClient)
	restClient.On("JobStart", mock.MatchedBy(func(in *restapi.JobStartIn) bool {
		return in.Job == "job" && *in.Command == "sh ../../tests/fixtures/echo.sh 3 5"
	})).Return(uuid.New(), nil)

	outFile := createTestOutFile(t, "out_")
	defer func() {
		outFile.Close()
		os.Remove(outFile.Name())
	}()
	cmdChan := make(chan *exec.Cmd, 1)
	exectr := executor.NewExecutor(restClient, executor.WithOutFile(outFile), executor.WithCmdChan(cmdChan))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		proc := <-cmdChan
		time.Sleep(time.Second)
		proc.Process.Signal(os.Interrupt)
		wg.Done()
	}()
	code, err := exectr.StartAndWatch(context.Background(), "job", []string{"sh", "../../tests/fixtures/echo.sh", "3", "5"})
	assert.Equal(t, 0, code)
	assert.NoError(t, err)
	wg.Wait()

	realOutput := readOutFile(t, outFile)
	assert.Equal(t, "SIGINT\n", realOutput[len(realOutput)-7:])
}
