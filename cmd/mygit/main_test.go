package main

import (
	"os"
	"testing"
  "bytes"
  "io"
  "compress/zlib"
)

func TestIsValidSHA1(t *testing.T) {
  cases := []struct {
    input          string
    expectedOutput bool
  }{
    {"", false},
    {"a", false},
    {"addf120b430021c36c232c99ef8d926aea2ac世", false},
    {"addf120b430021c36c232c99ef8d926aea2acd6世", false},
    {"addf120b430021c36c232c99ef8d926aea2acd6b123456789", false},
    {"addf120b430021c36c232c99ef8d926aea2acd6b", true},
  }
  for _, c := range cases {
    if output := isValidSHA1(c.input); output != c.expectedOutput {
        t.Errorf("incorrect output for `%s`: expected `%t` but got `%t`", c.input, c.expectedOutput, output)
    }
  }
}

// captureOutput 
func captureOutput() (*bytes.Buffer, *bytes.Buffer, func()) {
    // Save original stdout and stderr
    originalStdout, originalStderr := os.Stdout, os.Stderr

    // Create pipes for stdout and stderr
    rOut, wOut, _ := os.Pipe()
    rErr, wErr, _ := os.Pipe()

    os.Stdout, os.Stderr = wOut, wErr

    // Buffers to capture the outputs
    var bufOut, bufErr bytes.Buffer

    // Cleanup function to restore original state and capture outputs
    cleanup := func() {
        // Close the write ends of the pipes
        wOut.Close()
        wErr.Close()

        // Copy the outputs from the pipes to the buffers
        io.Copy(&bufOut, rOut)
        io.Copy(&bufErr, rErr)

        // Close the read ends of the pipes
        rOut.Close()
        rErr.Close()

        // Restore original stdout and stderr
        os.Stdout, os.Stderr = originalStdout, originalStderr
    }

    return &bufOut, &bufErr, cleanup
}

func setupCatFileTestFiles(t *testing.T) {
  for _, dir := range []string{".git", ".git/objects", ".git/objects/a9"} {
    if err := os.MkdirAll(dir, 0755); err != nil {
      t.Fatalf("Failed to create test directory: %v", err)
    }
  }

  filePath := ".git/objects/a9/4a8fe5ccb19ba61c4c0873d391e987982fbbd3"
  f, err := os.Create(filePath)
  if err != nil {
    t.Fatalf("Failed to create test file: %v", err)
  }
  defer f.Close()

	zw := zlib.NewWriter(f)
	defer zw.Close() 

	_, err = zw.Write([]byte("header\x00if you're reading this it worked"))
	if err != nil {
    t.Fatalf("Failed to write data to test file: %v", err)
	}
}

func cleanupCatFileTestFiles() {
  os.Remove(".git/objects/a9/4a8fe5ccb19ba61c4c0873d391e987982fbbd3")
}

func TestCatFile(t *testing.T) {
  cases := []struct {
    args       []string
    wantStdout string
    wantStderr string
  }{
    {[]string{ "-p", "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"}, "if you're reading this it worked", ""},
    {[]string{"-p", "invalidsha1"}, "", "The provided hash could not be verified, please provide a valid SHA1 hash\n"},
    {[]string{"-p", "67385b86859e3265d93eaf38cad7d06533ac4998"}, "", "The object corresponding to the hash 67385b86859e3265d93eaf38cad7d06533ac4998 does not exist.\n"},
  }

  setupCatFileTestFiles(t)
  defer cleanupCatFileTestFiles()


  for _, c := range cases {
    t.Run(c.args[1], func(t *testing.T) {
      // Capture output
      stdout, stderr, restore := captureOutput()

      catFile(c.args)

      restore()

      if got := stdout.String(); got != c.wantStdout {
        t.Errorf("stdout = %q, want %q", got, c.wantStdout)
      }

      if got := stderr.String(); got != c.wantStderr {
        t.Errorf("stderr = %q, want %q", got, c.wantStderr)
      }
    })
  }
}

func cleanupHashObjectTestFiles() {
  os.Remove("test.txt")
}

func setupHashObjectTestFiles(t *testing.T) {
  for _, dir := range []string{".git", ".git/objects"} {
    if err := os.MkdirAll(dir, 0755); err != nil {
      t.Fatalf("Failed to create test directory: %v", err)
    }
  }

  filePath := "test.txt"
  f, err := os.Create(filePath)
  if err != nil {
    t.Fatalf("Failed to create test file: %v", err)
  }
  defer f.Close()

	_, err = f.Write([]byte("if you're reading this it worked"))
	if err != nil {
    t.Fatalf("Failed to write data to test file: %v", err)
	}
}

func TestHashFile(t *testing.T) {
  cases := []struct {
    args       []string
    wantStdout string
    wantStderr string
  }{
    {[]string{"-w", "test.txt"}, "c12ff9bfd17010b62e9041ad4a414b3d608471af", ""},
  }

  setupHashObjectTestFiles(t)
  defer cleanupHashObjectTestFiles()


  for _, c := range cases {
    t.Run(c.args[1], func(t *testing.T) {
      // Capture output
      stdout, stderr, restore := captureOutput()

      hashObject(c.args)

      restore()

      if got := stdout.String(); got != c.wantStdout {
        t.Errorf("stdout = %q, want %q", got, c.wantStdout)
      }

      if got := stderr.String(); got != c.wantStderr {
        t.Errorf("stderr = %q, want %q", got, c.wantStderr)
      }
    })
  }
}
