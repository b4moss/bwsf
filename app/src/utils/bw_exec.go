package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

// bwResult is the outcome of a bw (or LookPath-backed) process invocation.
type bwResult struct {
	Output []byte // CombinedOutput merge, or stdout when SeparateStreams/StdoutOnly
	Stderr []byte // only populated when SeparateStreams
	Err    error
}

// bwRunOptions controls stdin/env and how streams are captured.
type bwRunOptions struct {
	Stdin           []byte
	Env             []string // nil = process default env
	SeparateStreams bool     // stdout→Output, stderr→Stderr (cmd.Run)
	StdoutOnly      bool     // cmd.Output() semantics
}

type bwRunner func(name string, args []string, opts bwRunOptions) bwResult

var (
	lookPath = exec.LookPath
	runBw    = defaultRunBw
)

func defaultRunBw(name string, args []string, opts bwRunOptions) bwResult {
	cmd := exec.Command(name, args...)
	if len(opts.Stdin) > 0 {
		cmd.Stdin = bytes.NewReader(opts.Stdin)
	}
	if opts.Env != nil {
		cmd.Env = opts.Env
	}
	if opts.SeparateStreams {
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		return bwResult{Output: stdout.Bytes(), Stderr: stderr.Bytes(), Err: err}
	}
	if opts.StdoutOnly {
		out, err := cmd.Output()
		return bwResult{Output: out, Err: err}
	}
	out, err := cmd.CombinedOutput()
	return bwResult{Output: out, Err: err}
}

func requireBwInstalled() error {
	if _, err := lookPath("bw"); err != nil {
		return fmt.Errorf("bw command is not installed")
	}
	return nil
}

// TestBwCall describes a stubbed bw invocation for cross-package tests (#160).
type TestBwCall struct {
	Name  string
	Args  []string
	Stdin []byte
}

// TestBwReply is the stubbed result for a TestBwCall.
type TestBwReply struct {
	Output []byte
	Stderr []byte
	Err    error
}

// InstallBwExecTestHook replaces LookPath and runBw for tests outside package utils.
// Call the returned restore from t.Cleanup.
func InstallBwExecTestHook(
	look func(file string) (string, error),
	handler func(TestBwCall) TestBwReply,
) (restore func()) {
	origLook := lookPath
	origRun := runBw
	if look != nil {
		lookPath = look
	}
	if handler != nil {
		runBw = func(name string, args []string, opts bwRunOptions) bwResult {
			reply := handler(TestBwCall{Name: name, Args: args, Stdin: opts.Stdin})
			return bwResult{Output: reply.Output, Stderr: reply.Stderr, Err: reply.Err}
		}
	}
	return func() {
		lookPath = origLook
		runBw = origRun
	}
}
