// SPDX-FileCopyrightText: Copyright 2021 The subcommandsutil Authors
// SPDX-License-Identifier: BSD-3-Clause

package subcommandsutil_test

import (
	"context"
	"flag"
	"sync"
	"testing"
	"time"

	"github.com/google/subcommands"

	"github.com/zchee/subcommandsutil"
)

func TestCancelableExecute(t *testing.T) {
	tests := map[string]struct {
		// Whether to cancel the execution context early.
		cancelContextEarly bool

		// Whether the underlying subcommand is expected to finish
		expectToFinish bool
	}{
		"when context is canceled early": {
			cancelContextEarly: true,
			expectToFinish:     false,
		},
		"when context is never canceled": {
			cancelContextEarly: false,
			expectToFinish:     true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tcmd := &testCommand{}
			cmd := subcommandsutil.Cancelable(tcmd)
			ctx, cancel := context.WithCancel(context.Background())

			if tt.cancelContextEarly {
				cancel()
				cmd.Execute(ctx, flag.NewFlagSet("test", flag.ContinueOnError))
			} else {
				cmd.Execute(ctx, flag.NewFlagSet("test", flag.ContinueOnError))
				cancel()
			}

			switch {
			case tcmd.DidFinish() && !tt.expectToFinish:
				t.Fatal("wanted command to exit early but it finished")
			case !tcmd.DidFinish() && tt.expectToFinish:
				t.Fatal("wanted command to finish but it exited early")
			}
		})
	}
}

// TestCancelableDelegation verifies that Cancelable() returns a subcommand.Command that
// delegates to the input subcommand.Command.
func TestCancelableDelegation(t *testing.T) {
	expectEq := func(t *testing.T, name, expected, actual string) {
		if expected != actual {
			t.Fatalf("wanted %s to be %q but got %q", name, expected, actual)
		}
	}

	cmd := subcommandsutil.Cancelable(&testCommand{
		name:     "test_name",
		usage:    "test_usage",
		synopsis: "test_synopsis",
	})
	expectEq(t, "Name", "test_name", cmd.Name())
	expectEq(t, "Usage", "test_usage", cmd.Usage())
	expectEq(t, "Synopsis", "test_synopsis", cmd.Synopsis())
}

type testCommand struct {
	name        string
	usage       string
	synopsis    string
	didFinish   bool
	didFinishMu sync.RWMutex
}

func (tcmd *testCommand) Name() string             { return tcmd.name }
func (tcmd *testCommand) Usage() string            { return tcmd.usage }
func (tcmd *testCommand) Synopsis() string         { return tcmd.synopsis }
func (tcmd *testCommand) SetFlags(f *flag.FlagSet) {}
func (tcmd *testCommand) Dispose() error           { return nil }

func (tcmd *testCommand) DidFinish() bool {
	tcmd.didFinishMu.RLock()
	defer tcmd.didFinishMu.RUnlock()

	return tcmd.didFinish
}

func (tcmd *testCommand) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	time.Sleep(time.Millisecond)

	tcmd.didFinishMu.Lock()
	tcmd.didFinish = true
	tcmd.didFinishMu.Unlock()

	return subcommands.ExitSuccess
}
