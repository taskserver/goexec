package goexec

import (
	"context"
	"os"
	"testing"
)

func Test_ExecuteShellScript(t *testing.T) {
	// Create a temporary script file for testing
	scriptContent := `#!/bin/sh
echo "Hello, World!"
`
	scriptFile, err := os.CreateTemp("", "test_script_*.sh")
	if err != nil {
		t.Fatalf("Failed to create temporary script file: %v", err)
	}
	defer os.Remove(scriptFile.Name())

	// Write the script content to the file
	_, err = scriptFile.WriteString(scriptContent)
	if err != nil {
		t.Fatalf("Failed to write to temporary script file: %v", err)
	}

	// Make the script executable
	err = os.Chmod(scriptFile.Name(), 0755)
	if err != nil {
		t.Fatalf("Failed to make script executable: %v", err)
	}

	// Test execution with default shell
	ctx := context.Background()
	result, err := ExecuteShellScript(ctx, scriptFile.Name())
	if err != nil {
		t.Fatalf("Failed to execute script with default shell: %v", err)
	}

	if result.Stdout != "Hello, World!\n" {
		t.Errorf("Unexpected output with default shell. Got: %q, Expected: %q", result.Stdout, "Hello, World!\n")
	}

	// Test execution with custom shell (bash)
	result, err = ExecuteShellScript(ctx, scriptFile.Name(), WithShell("bash"))
	if err != nil {
		t.Fatalf("Failed to execute script with custom shell: %v", err)
	}

	if result.Stdout != "Hello, World!\n" {
		t.Errorf("Unexpected output with custom shell. Got: %q, Expected: %q", result.Stdout, "Hello, World!\n")
	}
}

func Test_ExecuteShellScriptWithEnv(t *testing.T) {
	// Create a temporary script file for testing
	scriptContent := `#!/bin/sh
echo "GREETING: $GREETING"
echo "FOO: $FOO"
`
	scriptFile, err := os.CreateTemp("", "test_script_with_env_*.sh")
	if err != nil {
		t.Fatalf("Failed to create temporary script file: %v", err)
	}
	defer os.Remove(scriptFile.Name())

	// Write the script content to the file
	_, err = scriptFile.WriteString(scriptContent)
	if err != nil {
		t.Fatalf("Failed to write to temporary script file: %v", err)
	}

	// Make the script executable
	err = os.Chmod(scriptFile.Name(), 0755)
	if err != nil {
		t.Fatalf("Failed to make script executable: %v", err)
	}

	// Test execution with custom environment variables
	ctx := context.Background()
	customEnv := []string{"GREETING=Hello, Custom World!", "FOO=BarValue"}
	result, err := ExecuteShellScript(ctx, scriptFile.Name(), WithEnv(customEnv))
	if err != nil {
		t.Fatalf("Failed to execute script with custom environment variables: %v", err)
	}

	// Expected output
	expectedOutput := "GREETING: Hello, Custom World!\nFOO: BarValue\n"
	if result.Stdout != expectedOutput {
		t.Errorf("Unexpected output with custom environment variables. Got: %q, Expected: %q", result.Stdout, expectedOutput)
	}
}

func Test_ExecuteShellScriptWithArgs(t *testing.T) {
	// Create a temporary script file for testing
	scriptContent := `#!/bin/sh
echo "First Argument: $1"
echo "Second Argument: $2"
`
	scriptFile, err := os.CreateTemp("", "test_script_with_args_*.sh")
	if err != nil {
		t.Fatalf("Failed to create temporary script file: %v", err)
	}
	defer os.Remove(scriptFile.Name())

	// Write the script content to the file
	_, err = scriptFile.WriteString(scriptContent)
	if err != nil {
		t.Fatalf("Failed to write to temporary script file: %v", err)
	}

	// Make the script executable
	err = os.Chmod(scriptFile.Name(), 0755)
	if err != nil {
		t.Fatalf("Failed to make script executable: %v", err)
	}

	// Test execution with arguments
	ctx := context.Background()
	args := []string{"HelloArg", "WorldArg"}
	result, err := ExecuteShellScript(ctx, scriptFile.Name(), WithArgs(args))
	if err != nil {
		t.Fatalf("Failed to execute script with arguments: %v", err)
	}

	// Expected output
	expectedOutput := "First Argument: HelloArg\nSecond Argument: WorldArg\n"
	if result.Stdout != expectedOutput {
		t.Errorf("Unexpected output with arguments. Got: %q, Expected: %q", result.Stdout, expectedOutput)
	}
}

func Test_ExecuteShellScriptWithCwd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_cwd_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary file within the directory
	tempFilePath := tempDir + "/testfile.txt"
	err = os.WriteFile(tempFilePath, []byte("Content from CWD test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary file in directory: %v", err)
	}

	// Create a temporary script file for testing
	scriptContent := `#!/bin/sh
if [ -f "testfile.txt" ]; then
  echo "File exists in CWD"
else
  echo "File not found in CWD"
fi
`
	scriptFile, err := os.CreateTemp("", "test_script_with_cwd_*.sh")
	if err != nil {
		t.Fatalf("Failed to create temporary script file: %v", err)
	}
	defer os.Remove(scriptFile.Name())

	// Write the script content to the file
	_, err = scriptFile.WriteString(scriptContent)
	if err != nil {
		t.Fatalf("Failed to write to temporary script file: %v", err)
	}

	// Make the script executable
	err = os.Chmod(scriptFile.Name(), 0755)
	if err != nil {
		t.Fatalf("Failed to make script executable: %v", err)
	}

	// Test execution with CWD set to tempDir
	ctx := context.Background()
	result, err := ExecuteShellScript(ctx, scriptFile.Name(), WithCwd(tempDir))
	if err != nil {
		t.Fatalf("Failed to execute script with custom CWD: %v", err)
	}

	// Expected output
	expectedOutput := "File exists in CWD\n"
	if result.Stdout != expectedOutput {
		t.Errorf("Unexpected output with custom CWD. Got: %q, Expected: %q", result.Stdout, expectedOutput)
	}
}
