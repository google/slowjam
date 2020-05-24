### example 

### first add these lines to your main.go

```
package main

import (
    "github.com/google/slowjam/pkg/stacklog"
    "github.com/pkg/profile"
)

func main() {
		p := profile.Start(profile.TraceProfile, profile.NoShutdownHook)
		defer p.Stop()
		s, err := stacklog.Start(stacklog.Config{Path: "stack.log", Poll: 50 * time.Millisecond})
		if err != nil {
			panic("unable to log stacks")
		}
		defer s.Stop()
}
```

after runing your code. there will be a stack.log generated.

### convert stack.log to html

- install timeline2html
- run timeline2html

```
./timeline2html ./stack.log

```
this will generate two files slowjan_full.html and slowjam_simple.html