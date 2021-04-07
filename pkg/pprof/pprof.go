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
	"fmt"
	"time"

	"github.com/google/slowjam/pkg/stackparse"
	"google.golang.org/protobuf/proto"
	"k8s.io/klog/v2"
)

// ix returns the index of a label.
func ix(m map[string]int64, key string) int64 {
	i, ok := m[key]
	if ok {
		return i
	}

	m[key] = int64(len(m) + 1)

	return m[key]
}

// Render outputs a pprof protobuf somewhere.
func Render(samples []*stackparse.StackSample, ignoreCreators []string, goroutines []int) ([]byte, error) {
	st := map[string]int64{"": 0}

	p := &Profile{
		SampleType: []*ValueType{
			{Type: ix(st, "samples"), Unit: ix(st, "count")},
			{Type: ix(st, "latency"), Unit: ix(st, "nanoseconds")},
		},
		TimeNanos: time.Now().UnixNano(),
	}

	pss, loc, fx := processSamples(samples, st, ignoreCreators, goroutines)
	p.Sample = pss
	p.Location = loc
	p.Function = fx

	p.StringTable = make([]string, len(st)+1)
	for k, v := range st {
		p.StringTable[v] = k
	}

	return proto.Marshal(p)
}

func processSamples(samples []*stackparse.StackSample, st map[string]int64, ignoreCreators []string, goroutines []int) ([]*Sample, []*Location, []*Function) {
	ig := map[string]bool{}
	for _, i := range ignoreCreators {
		ig[i] = true
	}

	gorm := map[int]bool{}
	for _, i := range goroutines {
		gorm[i] = true
	}

	pss := []*Sample{}
	fmap := map[uint64]*Function{}
	lmap := map[uint64]*Location{}
	ftable := map[string]int64{}
	ltable := map[string]int64{}
	lastTime := samples[0].Time

	for _, s := range samples {
		locs := []uint64{}

		for _, g := range s.Context.Goroutines {
			if ig[g.CreatedBy.Func.PkgDotName()] {
				continue
			}

			if len(gorm) > 0 && !gorm[g.ID] {
				continue
			}

			for _, c := range g.Signature.Stack.Calls {
				if stackparse.InternalCall(c) {
					continue
				}

				f := &Function{
					Id:         uint64(ix(ftable, c.Func.Raw)),
					Name:       ix(st, c.Func.PkgDotName()),
					SystemName: ix(st, c.Func.PkgDotName()),
					Filename:   ix(st, c.SrcPath),
				}

				l := &Location{
					Id: uint64(ix(ltable, fmt.Sprintf("%s:%d", c.SrcPath, c.Line))),
					Line: []*Line{
						{FunctionId: f.Id, Line: int64(c.Line)},
					},
				}
				locs = append(locs, l.Id)
				fmap[f.Id] = f
				lmap[l.Id] = l
			}
		}

		if len(locs) == 0 {
			klog.Errorf("invalid sample, skipping")
			continue
		}

		pss = append(pss, &Sample{
			LocationId: locs,
			Value:      []int64{1, s.Time.Sub(lastTime).Nanoseconds()},
		})
		lastTime = s.Time
	}

	loc := []*Location{}
	for _, v := range lmap {
		loc = append(loc, v)
	}

	fx := []*Function{}
	for _, v := range fmap {
		fx = append(fx, v)
	}

	return pss, loc, fx
}
