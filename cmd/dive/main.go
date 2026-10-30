package main

// Copyright © 2018 Alex Goodman
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/anchore/clio"

	"github.com/wagoodman/dive/cmd/dive/cli"
	"github.com/wagoodman/dive/internal/source"
)

// applicationName is the non-capitalized name of the application (do not change this)
const (
	applicationName = "dive"
	notProvided     = "[not provided]"
)

// TODO: these need to be wired up to the build flags
// all variables here are provided as build-time arguments, with clear default values
var (
	version        = notProvided
	buildDate      = notProvided
	gitCommit      = notProvided
	gitDescription = notProvided
)

func main() {
	// stereoscope sends os.Args[0] as the registry User-Agent, keep the local install path (and username) out of it.
	// Nothing else reads more than the base name.
	os.Args[0] = filepath.Base(os.Args[0])

	app := cli.Application(
		clio.Identification{
			Name:           applicationName,
			Version:        version,
			BuildDate:      buildDate,
			GitCommit:      gitCommit,
			GitDescription: gitDescription,
		},
	)

	app.Run()

	// on an interrupt clio returns here without waiting for the command, which then never cleans up its image.
	// (on an error exit clio calls os.Exit instead, but the command has already finished and cleaned up then). A
	// second interrupt while Run is still running makes its deferred os.Exit(1) skip this, which can leave temp dirs
	// behind. clio has stopped listening for signals by now, so ignore them for the bounded wait, otherwise another
	// ctrl-c would kill dive mid cleanup.
	signal.Ignore(os.Interrupt)
	if err := source.Cleanup(5 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "unable to clean up temp dirs: %v\n", err)
	}
}
