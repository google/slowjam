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
	"github.com/google/slowjam/pkg/stackparse"
	"os"
	"strings"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("open: %v", err))
	}
	defer f.Close()
	samples, err := stackparse.Read(f)
	if err != nil {
		panic(fmt.Sprintf("parse: %v", err))
	}

	tl := stackparse.CreateTimeline(samples, stackparse.SuggestedIgnore)

	fmt.Printf("%d samples over %s\n", tl.Samples, tl.End.Sub(tl.Start))
	for _, g := range tl.Goroutines {
		fmt.Printf("goroutine %d (%s)\n", g.ID, g.Signature.CreatedByString(true))
		for i, l := range g.Layers {
			for _, c := range l.Calls {
				if c.Samples > 1 {
					fmt.Printf(" %s %s execution time: %s (%d samples)\n", strings.Repeat(" ", i), c.Name, c.EndDelta-c.StartDelta, c.Samples)
				}
			}
		}
		fmt.Printf("\n")
	}
}
