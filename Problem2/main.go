package main                       //Declares the main package. A Go program must always have a main package to be executable.  

import "fmt"                       //Imports the 'fmt' package, which provides functions for formatted I/O, including printing to the standard output

func main() {
    cnp := make(chan func(), 10)   //Creates a buffered channel cnp that can hold functions with no arguments or return values. The channel has a buffer capacity of 10
    for i := 0; i < 4; i++ {       //A loop that runs four times
        go func() {				   //Starts a new goroutine with an anonymous function. Each goroutine listens on the cnp channel for incoming functions and executes them
            for f := range cnp {   //Inside each goroutine, it uses a range loop to receive functions from the cnp channel and execute them
                f()
            }
        }()
    }
    cnp <- func() {                //Sends an anonymous function that prints "HERE1" to the cnp channel.
        fmt.Println("HERE1")
    }
    fmt.Println("Hello")		   // Prints "Hello"
}

// The code creates a buffered channel to send functions for execution to multiple goroutines. It then sends a function that prints "HERE1" to the channel and prints "Hello" separately.

//  in this code snippet the output will be 'Hello'