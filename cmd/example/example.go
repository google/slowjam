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
	"context"
	"fmt"
	"math"
	"os/exec"
	"runtime/trace"
	"time"

	"github.com/pkg/profile"
	"github.com/google/slowjam/pkg/stacklog"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	p := profile.Start(profile.TraceProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	defer p.Stop()
	s, err := stacklog.Start(stacklog.Config{Path: "stack.log", Poll: 50 * time.Millisecond})
	if err != nil {
		panic("unable to log stacks")
	}
	defer s.Stop()

	fmt.Println("start")
	goToSleep(ctx)
	var g errgroup.Group
	g.Go(func() error {
		time.Sleep(5 * time.Second)
		fmt.Println("errgroup end")
		return nil
	})
	fmt.Println("pi")
	calcPI()
	runCmd()
	runCmd()
	fmt.Println("wait")
	waitPlease(&g)
	fmt.Println("end")
}

func calcPI() {
	cosVal := float64(-1)
	for n := 4; n < 50000000; n *= 2 {
		cosVal = math.Sqrt(0.5 * (cosVal + 1.0))
		math.Pow(0.5-0.5*cosVal, 0.5)
	}
}

func runCmd() {
	out, err := exec.Command("iostat", "1", "2").CombinedOutput()
	fmt.Printf("out: %s err: %v\n", out, err)
}

func goToSleep(ctx context.Context) {
	ctx, task := trace.NewTask(ctx, "Sleep task")
	time.Sleep(500 * time.Millisecond)
	task.End()
}

func waitPlease(eg *errgroup.Group) {
	eg.Wait()
}
