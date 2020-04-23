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
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/golang/glog"
	"github.com/maruel/panicparse/stack"
	"github.com/pkg/browser"
	"golang.org/x/image/colornames"

	"github.com/google/slowjam/pkg/stackparse"
)

var (
	port     = flag.Int("port", 8000, "service port")
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
	http.HandleFunc("/simple", DisplayTimeline(tls))
	http.HandleFunc("/", DisplayTimeline(tl))

	listenAddr := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if listenAddr == ":" {
		listenAddr = fmt.Sprintf(":%d", *port)
	}
	url := fmt.Sprintf("http://localhost%s/", listenAddr)
	fmt.Printf("Opening %s ...", url)
	go func() {
		time.Sleep(1)
		browser.OpenURL(url)
	}()
	err = http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic(err)
	}
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

// Display a timeline
func DisplayTimeline(tl *stackparse.Timeline) http.HandlerFunc {
	updateColorMap(tl, colorMap)
	fmap := template.FuncMap{
		"Milliseconds": milliseconds,
		"Creator":      creator,
		"Height":       height,
		"Color":        callColor,
	}
	t := template.Must(template.New("timeline").Funcs(fmap).ParseFiles("timeline.tmpl"))

	return func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("Timeline request coming in")
		rc := struct {
			Duration time.Duration
			TL       *stackparse.Timeline
		}{
			Duration: tl.End.Sub(tl.Start),
			TL:       tl,
		}
		err := t.ExecuteTemplate(w, "timeline.tmpl", rc)
		if err != nil {
			glog.Errorf("tmpl: %v", err)
			return
		}
	}
}
