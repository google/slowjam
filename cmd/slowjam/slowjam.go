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
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"github.com/google/slowjam/pkg/pprof"
	"github.com/google/slowjam/pkg/stacklog"
	"github.com/google/slowjam/pkg/stackparse"
	"github.com/google/slowjam/pkg/text"
	"github.com/google/slowjam/pkg/web"
)

var (
	httpEndpoint = pflag.String("http", "", "HTTP endpoint to listen at")
	htmlPath     = pflag.String("html", "", "Path to output HTML content to")
	pprofPath    = pflag.String("pprof", "", "Path to output pprof content to (consider using --goroutines=1)")
	goroutines   = pflag.IntSlice("goroutines", []int{}, "goroutines to include (default: all)")
	dumpText     = pflag.Bool("text", false, "Outputs text rendering of goroutines found")
)

func main() {
	s := stacklog.MustStartFromEnv("STACKLOG_PATH")
	defer s.Stop()

	pflag.Parse()

	if len(pflag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "usage: slowjam [flags] <path>")
		os.Exit(64) // EX_USAGE
	}

	f, err := os.Open(pflag.Args()[0])
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

	tl := stackparse.CreateTimeline(samples, stackparse.SuggestedIgnore, *goroutines)

	if *httpEndpoint != "" {
		web.Serve(*httpEndpoint, tl)
		return
	}

	if *htmlPath != "" {
		w, err := os.Create(*htmlPath)
		if err != nil {
			glog.Exitf("open failed: %v", err)
		}
		defer w.Close()

		if err := web.Render(w, tl); err != nil {
			glog.Fatalf("render: %v", err)
		}

		return
	}

	if *pprofPath != "" {
		w, err := os.Create(*pprofPath)
		if err != nil {
			glog.Exitf("open failed: %v", err)
		}
		defer w.Close()

		bs, err := pprof.Render(samples, stackparse.SuggestedIgnore, *goroutines)
		if err != nil {
			glog.Fatalf("render: %v", err)
		}

		if _, err := w.Write(bs); err != nil {
			glog.Fatalf("write: %v", err)
		}

		return
	}

	if *dumpText {
		fmt.Print(text.Tree(tl))
		return
	}

	glog.Exitf("no output mode specified")
}
