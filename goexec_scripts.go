package goexec

import (
	"context"
	"fmt"
	"os"
)

type ShellScriptOption func(*ExecTask)

func WithEnv(env []string) ShellScriptOption {
	return func(et *ExecTask) {
		et.Env = env
	}
}

func WithShell(shell string) ShellScriptOption {
	return func(et *ExecTask) {
		if shell != "" {
			et.Command = shell
			et.Shell = true
		}
	}
}

func WithCwd(cwd string) ShellScriptOption {
	return func(et *ExecTask) {
		et.Cwd = cwd
	}
}

func WithArgs(args []string) ShellScriptOption {
	return func(et *ExecTask) {
		et.Args = append(et.Args, args...)
	}
}

func WithOutputFiles(stdout, stderr *os.File) ShellScriptOption {
	return func(et *ExecTask) {
		et.OutputFile = stdout
		et.ErrorFile = stderr
	}
}

func ExecuteShellScript(ctx context.Context, scriptPath string, opts ...ShellScriptOption) (ExecResult, error) {
	// Check if the script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return ExecResult{}, fmt.Errorf("script not found: %s", scriptPath)
	}

	// Default ExecTask setup
	execTask := ExecTask{
		Command: "sh", // Default shell
		Args:    []string{scriptPath},
		Shell:   true, // Ensure it runs in a shell
	}

	// Apply options (will override defaults if provided)
	for _, opt := range opts {
		opt(&execTask)
	}

	// Execute the task
	return execTask.Execute(ctx)
}
