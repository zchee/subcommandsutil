// SPDX-FileCopyrightText: Copyright 2021 The subcommandsutil Authors
// SPDX-License-Identifier: BSD-3-Clause

package subcommandsutil

import (
	"context"
	"flag"
	"log"
	"runtime"

	"github.com/google/subcommands"
)

// CancelableCommand is an object that performs tear down. This is used by Cancelable to gracefully
// terminate a delegate Command before exiting.
type CancelableCommand interface {
	subcommands.Command

	// Dispose provides the gracefully terminate a delegate Command before exiting.
	Dispose() error
}

// cancelable wraps a subcommands.Command so that it is canceled if the input execution
// context emits a Done event before execution is finished. cancelable "masquerades" as
// the underlying Command. Example Registration:
//
//   subcommands.Register(subcommandsutil.Cancelable(&OtherSubcommand{}))
type cancelable struct {
	sub CancelableCommand
}

// make sure cancelable implements the subcommands.Command interface.
var _ subcommands.Command = (*cancelable)(nil)

// Cancelable wraps a subcommands.Command so that it is canceled if its input execution
// context emits a Done event before execution is finished.
//
// The wrapped sub will calling Dispose before the program exits.
func Cancelable(sub CancelableCommand) subcommands.Command {
	return &cancelable{
		sub: sub,
	}
}

// Name forwards to the underlying c.sub Command.
func (c *cancelable) Name() string {
	return c.sub.Name()
}

// Usage forwards to the underlying c.sub Command.
func (c *cancelable) Usage() string {
	return c.sub.Usage()
}

// Synopsis forwards to the underlying c.sub Command.
func (c *cancelable) Synopsis() string {
	return c.sub.Synopsis()
}

// SetFlags forwards to the underlying c.sub Command.
func (c *cancelable) SetFlags(f *flag.FlagSet) {
	c.sub.SetFlags(f)
}

// Execute runs the underlying Command in a goroutine.
//
// If the input context is canceled before execution finishes, execution is canceled and the context's error is logged.
func (c *cancelable) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	ch := make(chan subcommands.ExitStatus)
	go func() {
		defer runtime.Goexit()
		ch <- c.sub.Execute(ctx, f, args...)
	}()

	select {
	case <-ctx.Done():
		_ = c.sub.Dispose()    // TODO(zchee): hasdling error
		log.Println(ctx.Err()) // TODO(zchee): use custom logger
		return subcommands.ExitFailure

	case s := <-ch:
		close(ch)
		return s
	}
}
