/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"

	"github.com/google/slowjam/pkg/stacklog"
	"github.com/google/slowjam/pkg/stackparse"
	"github.com/google/slowjam/pkg/text"
	"github.com/google/slowjam/pkg/web"
)

var (
	httpEndpoint = flag.String("http", "", "HTTP endpoint to listen at")
	htmlPath     = flag.String("html", "", "HTML path to output to")
	dumpText     = flag.Bool("text", false, "Outputs text rendering of goroutines found")
)

func main() {
	s := stacklog.MustStartFromEnv("STACKLOG_PATH")
	defer s.Stop()

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "usage: slowjam [flags] <path>")
		os.Exit(64) // EX_USAGE
	}

	f, err := os.Open(flag.Args()[0])
	if err != nil {
		glog.Fatalf("open: %v", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			glog.Errorf("close failed: %v", err)
		}
	}()

	samples, err := stackparse.Read(f)
	if err != nil {
		glog.Fatalf("parse: %v", err)
	}

	tl := stackparse.CreateTimeline(samples, stackparse.SuggestedIgnore)

	if *httpEndpoint != "" {
		web.Serve(*httpEndpoint, tl)
		return
	}

	if *htmlPath != "" {
		w, err := os.Open(*htmlPath)
		if err != nil {
			glog.Exitf("open failed: %v", err)
		}

		if err := web.Render(w, tl); err != nil {
			glog.Fatalf("render: %v", err)
		}

		return
	}

	if *dumpText {
		fmt.Print(text.Tree(tl))
		return
	}

	glog.Exitf("no output mode specified")
}
