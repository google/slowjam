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

// Package text is for rendering a timeline into text form
package text

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/slowjam/pkg/stackparse"
)

// Tree outputs a human-readable tree of goroutines found.
func Tree(tl *stackparse.Timeline) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%d samples over %s\n", tl.Samples, tl.End.Sub(tl.Start)))

	gor := []int{}

	for k := range tl.Goroutines {
		gor = append(gor, k)
	}

	sort.Ints(gor)

	for _, gid := range gor {
		g := tl.Goroutines[gid]
		funcName := ""
		if len(g.Signature.CreatedBy.Calls) != 0 {
			call := g.Signature.CreatedBy.Calls[0]
			funcName = fmt.Sprintf("%s @ %s:%d", stackparse.PkgDotName(call.Func), call.RemoteSrcPath, call.Line)
		}
		sb.WriteString(fmt.Sprintf("goroutine %d (%s)\n", gid, funcName))

		for i, l := range g.Layers {
			for _, c := range l.Calls {
				if c.Samples > 1 {
					sb.WriteString(fmt.Sprintf(" %s %s execution time: %s (%d samples)\n", strings.Repeat(" ", i), c.Name, c.EndDelta-c.StartDelta, c.Samples))
				}
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
