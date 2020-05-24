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
	"image/color"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/maruel/panicparse/stack"
	"golang.org/x/image/colornames"

	"github.com/google/slowjam/pkg/stackparse"
)

var (
	colorMap = map[string]color.RGBA{}
)

func main() {
	flag.Parse()

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
	tls := stackparse.SimplifyTimeline(tl)
	GenerateHTML(tl, "full.html")
	GenerateHTML(tls, "simple.html")

}

func milliseconds(d time.Duration) string {
	return fmt.Sprintf("%d", d.Milliseconds())
}

func creator(s *stack.Signature) string {
	c := s.CreatedBy.Func.PkgDotName()
	if c == "" {
		c = "main"
	}
	return c
}

func height(ls []*stackparse.Layer) string {
	return fmt.Sprintf("%d", 100+(35*len(ls)))
}

func updateColorMap(tl *stackparse.Timeline, cm map[string]color.RGBA) {
	chosen := map[string]bool{}

	for _, g := range tl.Goroutines {
		for _, l := range g.Layers {
			for _, c := range l.Calls {
				_, ok := cm[c.Package]
				if ok {
					continue
				}

				// gimmick: try to find a similar name
				for name, value := range colornames.Map {
					if strings.Contains(name, "white") {
						continue
					}
					if name[0] != c.Package[0] {
						continue
					}

					if !chosen[name] {
						chosen[name] = true
						cm[c.Package] = value
						break
					}
				}

				_, ok = cm[c.Package]
				if ok {
					continue
				}

				// Giveup
				for name, value := range colornames.Map {
					if strings.Contains(name, "white") {
						continue
					}
					if !chosen[name] {
						chosen[name] = true
						cm[c.Package] = value
						break
					}
				}
			}
		}
	}
}

func callColor(pkg string, level int) string {
	c := colorMap[pkg]
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// GenerateHTML generates an HTML file out of the stackparse.Timeline
func GenerateHTML(tl *stackparse.Timeline, fileName string) {
	updateColorMap(tl, colorMap)

	glog.Infof("Timeline request coming in")

	rc := struct {
		Duration time.Duration
		TL       *stackparse.Timeline
	}{
		Duration: tl.End.Sub(tl.Start),
		TL:       tl,
	}

	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		panic(fmt.Sprintf("open: %v", err))
	}

	err = reportTemplate.Execute(f, rc)
	if err != nil {
		panic(fmt.Sprintf("execute html: %v", err))
	}

}
