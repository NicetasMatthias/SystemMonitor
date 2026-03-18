# SystemMonitor
System monitoring daemon with web server


## Usage

`make all`	    Default target. Downloads dependencies and builds the binary.
`make build`	Downloads dependencies and compiles the source code. The output binary is placed in the bin/ directory.
`make run`	    Builds the project and immediately runs the resulting binary.
`make test`	    Runs all tests with verbose output and the race detector enabled.
`make deps`	    Downloads and verifies the Go module dependencies.
`make clean`	Removes the bin/ directory and cleans up the Go build cache.
`make help`	    Displays a list of available Makefile commands with descriptions.