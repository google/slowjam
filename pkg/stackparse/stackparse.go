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

// Package stackparse turns stacklogs into objects for analysis
package stackparse

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/maruel/panicparse/stack"
)

// StackSample represents a single Go stack at a point in time.
type StackSample struct {
	Time    time.Time
	Context *stack.Context
}

// Read parses a stack log input.
func Read(r io.Reader) ([]*StackSample, error) {
	inStack := false
	t := time.Time{}
	sd := bytes.NewBuffer([]byte{})
	samples := []*StackSample{}

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		if !inStack {
			line := scanner.Text()

			s, err := strconv.ParseInt(line, 10, 64)
			if err != nil {
				return samples, err
			}

			t = time.Unix(0, s)
			inStack = true

			continue
		}

		if strings.HasPrefix(scanner.Text(), "-") {
			inStack = false

			ctx, err := stack.ParseDump(sd, os.Stdout, false)
			if err != nil {
				return samples, err
			}

			samples = append(samples, &StackSample{Time: t, Context: ctx})

			continue
		}

		sd.Write(scanner.Bytes())
		sd.Write([]byte{'\n'})
	}

	if err := scanner.Err(); err != nil {
		return samples, err
	}

	return samples, nil
}
