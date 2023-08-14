# GoPie
This is the repo of Go-pie, a concurrency testing tool for Golang.
This document is to introduce the structure of `GoPie` project.

## Project Structure
- bug.md: a file used to track the bugs found by `GoPie` reported on Github. 
- cmd: files under `cmd` are the command line tools of `GoPie`, including the instrument tools and the testing tools. 
- patch: files under `patch` are the hacks of the runtime of Golang, these files will be replaced before compiling the target binary.
- pkg: the packages of GoPie, contain all the details in the form of source code. The important packages under `pkg` are:
  - feedback: the runtime feedback analysis in GoPie.
  - hack: the hack of Go runtime data structure.
  - inst: the passes of instrumentation.
  - sched: implement of scheduling approach.
- script: shell scripts used during development.
- Dockerfile: dev environment.

## Usage
1. `GoPie` has been implemented using `Go 1.19.1`. 

    Follow https://go.dev/doc/install to install the right version of Go.
2. Build the binary under `cmd` with `go build -o ./bin ./cmd/...`. There will be two binaries after compilation, the `inst` and `fuzz`. 
    ~~~shell
    go build -o ./bin ./cmd/...
    ~~~
2. Instrument stubs and compile your project which is to be tested.
    ~~~shell
    ./bin/fuzz --task inst --path your_project_to_be_tested
    ~~~
3. Build test binaries, the test binaries will be placed into `./testbins`
    ~~~shell
    // install dependencies
    cd your_project_to_be_tested
    go mod tidy

    // compile the unit tests
    ./bin/fuzz --task bins --path your_project_to_be_tested
    ~~~
4. Start testing
    ~~~
    ./bin/fuzz --task full --path path_of_test_binaries
    ~~~
