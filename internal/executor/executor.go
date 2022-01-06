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
	ExitError = 1
	ExitOK    = 0
)

var errInvalidArguments = errors.New("invalid arguments")

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
		return ExitError, fmt.Errorf("StartAndWatch: %w: command name is required", errInvalidArguments)
	}
	hostname, err := os.Hostname()
	if err != nil {
		return ExitError, fmt.Errorf("StartAndWatch: %w", err)
	}

	startIn := &restapi.JobStartIn{
		Job:       job,
		StartedAt: internal.NewPointerOfTime(time.Now()),
		Command:   internal.NewPointerOfString(strings.Join(args, " ")),
		Pid:       internal.NewPointerOfInt(os.Getpid()),
		Host:      &hostname,
	}

	_, err = e.client.JobStart(ctx, startIn)
	if err != nil {
		return ExitError, fmt.Errorf("send job start to api: %w", err)
	}

	var cmd *exec.Cmd
	if len(args) == 1 {
		cmd = exec.Command(args[0]) //nolint:gosec
	} else {
		cmd = exec.Command(args[0], args[1:]...) //nolint:gosec
	}

	cmd.Stdout = e.outFile
	cmd.Stderr = e.errFile

	if err := cmd.Start(); err != nil {
		return ExitError, fmt.Errorf("error start command: %w", err)
	}
	if e.cmdChan != nil {
		e.cmdChan <- cmd
	}

	return e.watch(ctx, cmd)
}

func (e *Executor) watch(ctx context.Context, cmd *exec.Cmd) (exitCode int, err error) {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	errs := make(chan error)

	wgCmd := sync.WaitGroup{}

	wgCmd.Add(1)
	go func() {
		defer wgCmd.Done()

		select {
		case <-ctx.Done():
			// TODO: parent process was finished
			return
		case <-done:
			// TODO: user finished process
			return
		case sig := <-sigs:
			if err := cmd.Process.Signal(sig); err != nil {
				errs <- fmt.Errorf("error sending signal to process: %w", err)

				return
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return ExitError, fmt.Errorf("error running command: %w", err)
	}
	done <- true

	wgCmd.Wait()

	return ExitOK, nil
}
