# go-markov
Simple Markov chain input and output. :)

Get package.
```go
go get "github.com/ActuallyGiggles/go-markov"
```

```go
package main

import (
	"fmt"
	"log"
	
	"github.com/ActuallyGiggles/go-markov"
)

func main() {
	i := markov.StartInstructions{
		WriteMode:  "counter",
		WriteLimit: 10000,
		StartKey:   "b5G(n1$I!4g",
		EndKey:     "e1$D(n7",
	}

	markov.Start(i)

	markov.Input("test", "This is a test.")

	oi := markov.OutputInstructions{
			Method: "TargetedBeginning",
			Chain:  "test",
			Target: "This",
	}

	output, err := markov.Output(oi)
	
	if err != nil {
		log.Println(err)
	}
	
	fmt.Println(output)
}
