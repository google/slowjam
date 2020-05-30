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

package web

import (
	"fmt"
	"image/color"
	"net/http"

	"github.com/google/slowjam/pkg/stackparse"
)

var (
	colorMap = map[string]color.RGBA{}
)

// Serve starts up an HTTP server at a given endpoint.
func Serve(endpoint string, tl *stackparse.Timeline) {
	tls := stackparse.SimplifyTimeline(tl)
	http.HandleFunc("/simple", displayTimeline(tls))
	http.HandleFunc("/", displayTimeline(tl))

	fmt.Printf("Listening at %s ...", endpoint)

	err := http.ListenAndServe(endpoint, nil)
	if err != nil {
		panic(err)
	}
}

func displayTimeline(tl *stackparse.Timeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := Render(w, tl); err != nil {
			http.Error(w, fmt.Sprintf("render failed: %v", err), 500)
		}
	}
}
