package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/antgubarev/pet/internal"
	"github.com/antgubarev/pet/internal/restapi"
)

const (
	EXIT_ERROR = 1
	EXIT_OK    = 0
)

type options struct {
	outFile *os.File
	errFile *os.File
	cmdChan chan *exec.Cmd
}

type Option func(*options)

func WithOutFile(out *os.File) Option {
	return func(o *options) {
		o.outFile = out
	}
}

func WithErrFile(err *os.File) Option {
	return func(o *options) {
		o.errFile = err
	}
}

func WithCmdChan(cmdChan chan *exec.Cmd) Option {
	return func(o *options) {
		o.cmdChan = cmdChan
	}
}

type Executor struct {
	options
	client restapi.Client
}

func NewExecutor(client restapi.Client, opts ...Option) *Executor {
	cli := &Executor{client: client}

	for _, optFunc := range opts {
		optFunc(&cli.options)
	}

	return cli
}

func (e *Executor) StartAndWatch(ctx context.Context, job string, args []string) (exitCode int, err error) {
	if len(args) == 0 {
		return EXIT_ERROR, errors.New("command name is required")
	}
	hostname, err := os.Hostname()
	if err != nil {
		return EXIT_ERROR, fmt.Errorf("executor start: %v", err)
	}

	startIn := &restapi.JobStartIn{
		Job:       job,
		StartedAt: internal.NewPointerOfTime(time.Now()),
		Command:   internal.NewPointerOfString(strings.Join(args, " ")),
		Pid:       internal.NewPointerOfInt(os.Getpid()),
		Host:      &hostname,
	}

	_, err = e.client.JobStart(startIn)
	if err != nil {
		return EXIT_ERROR, fmt.Errorf("send job start to api: %v", err)
	}

	var cmd *exec.Cmd
	if len(args) == 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	cmd.Stdout = e.outFile
	cmd.Stderr = e.errFile

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	errs := make(chan error)

	if err := cmd.Start(); err != nil {
		return EXIT_ERROR, fmt.Errorf("error start command: %v", err)
	}
	if e.cmdChan != nil {
		e.cmdChan <- cmd
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			// TODO: parent process finish
			return
		case <-done:
			// TODO: user finishs process
			return
		case sig := <-sigs:
			if err := cmd.Process.Signal(sig); err != nil {
				errs <- fmt.Errorf("error sending signal to process: %v", err)
				return
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return EXIT_ERROR, fmt.Errorf("error running command: %v", err)
	}
	done <- true

	wg.Wait()

	return EXIT_OK, nil
}
