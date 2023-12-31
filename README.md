# Pool Party
[![Go Reference](https://pkg.go.dev/badge/github.com/we-are-discussing-rest/pool-party.svg)](https://pkg.go.dev/github.com/we-are-discussing-rest/pool-party)

Super simple worker pools in Go

## Usage
Add the package
```shell
go get github.com/we-are-discussing-rest/pool-party
```
To create a new worker pool simply use `NewPool`

```go
import (
    poolparty "github.com/we-are-discussing-rest/pool-party"
)

func main() {
    p := poolparty.NewPool(5) // Creates a new worker pool with 5 workers

    // Note: Make sure you invoke the start() method before sending tasks to workers
	p.Start() // Starts all the workers in the pool
	
    // Note: this library only currently supports sending a function with no params
    p.Send(func() {
        time.Sleep(30 * time.Seconds)
    }) // Sends a new task to the worker pool 


    p.Stop() // Gracefully stops the workers 
}
```

