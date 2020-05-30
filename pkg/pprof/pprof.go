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

// Package pprof is for rendering a timeline into a pprof protobuf.
package pprof

import (
	"time"

	"github.com/google/slowjam/pkg/stackparse"
	"google.golang.org/protobuf/proto"
)

// ix returns the index of a label.
func ix(m map[string]int64, key string) int64 {
	i, ok := m[key]
	if ok {
		return i
	}

	m[key] = int64(len(m))

	return m[key]
}

// Render outputs a pprof protobuf somewhere.
func Render(tl *stackparse.Timeline) ([]byte, error) {
	m := map[string]int64{"": 0}

	p := &Profile{
		SampleType: []*ValueType{
			{Type: ix(m, "samples"), Unit: ix(m, "count")},
			{Type: ix(m, "latency"), Unit: ix(m, "nanoseconds")},
		},
		TimeNanos: time.Now().UnixNano(),
	}

	sts := make([]string, len(m))
	for k, v := range m {
		sts[v] = k
	}

	p.StringTable = sts

	return proto.Marshal(p)
}
