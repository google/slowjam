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

package stacklog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// Config defines how to configure a stack logger
type Config struct {
	Path string
	Poll time.Duration
}

// Start begins logging stacks to an output file
func Start(c Config) (*StackLog, error) {
	if c.Poll == 0 {
		c.Poll = 125 * time.Millisecond
	}
	if c.Path == "" {
		c.Path = "stack.log"
	}
	os.Stderr.WriteString(fmt.Sprintf("Logging stacks to %s, sampling every %s\n", c.Path, c.Poll))
	s := &StackLog{
		ticker: time.NewTicker(c.Poll),
		path:   c.Path,
	}
	f, err := os.Create(c.Path)
	if err != nil {
		return s, err
	}
	s.f = f
	go s.loop()
	return s, nil
}

// StackLog controls the stack logger
type StackLog struct {
	ticker  *time.Ticker
	f       io.WriteCloser
	path    string
	samples int
}

// loop starts a background
func (s *StackLog) loop() {
	for range s.ticker.C {
		s.f.Write([]byte(fmt.Sprintf("%d\n", time.Now().UnixNano())))
		s.f.Write(DumpStacks())
		s.f.Write([]byte("-\n"))
		s.samples++
	}
}

// DumpStacks returns a formatted stack trace of goroutines, using a large enough buffer to capture the entire trace
func DumpStacks() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

// Stop stops logging stacks
func (s *StackLog) Stop() {
	s.ticker.Stop()
	os.Stderr.WriteString(fmt.Sprintf("stacklog: disabled. stored %d samples to %s\n", s.samples, s.path))
}
