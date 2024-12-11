package goexec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ExecTask struct {
	// Command is the command to execute. This can be the path to an executable
	// or the executable with arguments. The arguments are detected by looking for
	// a space.
	//
	// Any arguments must be given via Args
	Command string

	// Args are the arguments to pass to the command. These are ignored if the
	// Command contains arguments.
	Args []string

	// Shell run the command in a bash shell.
	// Note that the system must have `bash` installed in the PATH or in /bin/bash
	Shell bool

	// Env is a list of environment variables to add to the current environment,
	// these are used to override any existing environment variables.
	Env []string

	// Cwd is the working directory for the command
	Cwd string

	// Stdin connect a reader to stdin for the command
	// being executed.
	Stdin io.Reader

	// PrintCommand prints the command before executing
	PrintCommand bool

	// StreamStdio prints stdout and stderr directly to os.Stdout/err as
	// the command runs.
	StreamStdio bool

	// DisableStdioBuffer prevents any output from being saved in the
	// TaskResult, which is useful for when the result is very large, or
	// when you want to stream the output to another writer exclusively.
	DisableStdioBuffer bool

	// StdoutWriter when set will receive a copy of stdout from the command
	StdOutWriter io.Writer

	// StderrWriter when set will receive a copy of stderr from the command
	StdErrWriter io.Writer

	// Timeout is the maximum time the command is allowed to run before being
	// cancelled and returning an error.
	Timeout time.Duration // Timeout for command execution

	OutputFile *os.File

	ErrorFile *os.File
}

type ExecResult struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Timedout  bool
	Cancelled bool
	Duration  time.Duration
}

func (et ExecTask) Execute(ctx context.Context) (ExecResult, error) {
	// If a timeout is set, create a new context with the timeout
	if et.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, et.Timeout)
		defer cancel()
	}
	argsSt := ""
	if len(et.Args) > 0 {
		argsSt = strings.Join(et.Args, " ")
	}

	if et.PrintCommand {
		fmt.Println("exec: ", et.Command, argsSt)
	}

	// don't try to run if the context is already cancelled
	if ctx.Err() != nil {
		return ExecResult{
			// the exec package returns -1 for cancelled commands
			ExitCode:  -1,
			Cancelled: ctx.Err() == context.Canceled,
		}, ctx.Err()
	}

	var command string
	var commandArgs []string
	if et.Shell {

		// On a NixOS system, /bin/bash doesn't exist at /bin/bash
		// the default behavior of exec.Command is to look for the
		// executable in PATH.

		command = "sh"
		// There is a chance that PATH is not populate or propagated, therefore
		// when bash cannot be resolved, set it to /bin/bash instead.
		if _, err := exec.LookPath(command); err != nil {
			command = "/usr/bin/sh"
		}

		if len(et.Args) == 0 {
			// use Split and Join to remove any extra whitespace?
			startArgs := strings.Split(et.Command, " ")
			script := strings.Join(startArgs, " ")
			commandArgs = append([]string{"-c"}, script)

		} else {
			script := strings.Join(et.Args, " ")
			commandArgs = append([]string{"-c"}, fmt.Sprintf("%s %s", et.Command, script))
		}
	} else {

		command = et.Command
		commandArgs = et.Args

		// AE: This had to be removed to fix: #117 where Windows users
		// have spaces in their paths, which are misinterpreted as
		// arguments for the command.
		// if strings.Contains(et.Command, " ") {
		// 	parts := strings.Split(et.Command, " ")
		// 	command = parts[0]
		// 	commandArgs = parts[1:]
		// }
	}

	cmd := exec.CommandContext(ctx, command, commandArgs...)
	cmd.Dir = et.Cwd

	if len(et.Env) > 0 {
		overrides := map[string]bool{}
		for _, env := range et.Env {
			key := strings.Split(env, "=")[0]
			overrides[key] = true
			cmd.Env = append(cmd.Env, env)
		}

		for _, env := range os.Environ() {
			key := strings.Split(env, "=")[0]

			if _, ok := overrides[key]; !ok {
				cmd.Env = append(cmd.Env, env)
			}
		}
	}
	if et.Stdin != nil {
		cmd.Stdin = et.Stdin
	}

	stdoutBuff := bytes.Buffer{}
	stderrBuff := bytes.Buffer{}

	var stdoutWriters []io.Writer
	var stderrWriters []io.Writer

	if et.OutputFile != nil {
		stdoutWriters = append(stdoutWriters, et.OutputFile)
	}

	if et.ErrorFile != nil {
		stderrWriters = append(stderrWriters, et.ErrorFile)
	}

	if !et.DisableStdioBuffer {
		stdoutWriters = append(stdoutWriters, &stdoutBuff)
		stderrWriters = append(stderrWriters, &stderrBuff)
	}

	if et.StreamStdio {
		stdoutWriters = append(stdoutWriters, os.Stdout)
		stderrWriters = append(stderrWriters, os.Stderr)
	}

	if et.StdOutWriter != nil {
		stdoutWriters = append(stdoutWriters, et.StdOutWriter)
	}
	if et.StdErrWriter != nil {
		stderrWriters = append(stderrWriters, et.StdErrWriter)
	}

	cmd.Stdout = io.MultiWriter(stdoutWriters...)
	cmd.Stderr = io.MultiWriter(stderrWriters...)
	execStart := time.Now()
	startErr := cmd.Start()
	if startErr != nil {
		return ExecResult{}, startErr
	}

	exitCode := 0
	execErr := cmd.Wait()
	execEnd := time.Now()
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return ExecResult{
		Stdout:    stdoutBuff.String(),
		Stderr:    stderrBuff.String(),
		ExitCode:  exitCode,
		Duration:  execEnd.Sub(execStart),
		Timedout:  ctx.Err() == context.DeadlineExceeded,
		Cancelled: ctx.Err() == context.Canceled,
	}, ctx.Err()
}
