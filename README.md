# prunejuice

*Violently ejecting the refuse*

A small CLI tool to delete old entries from inside a directory. These might be log files, config files you generate and churn through, you get the general idea. Keep recent ones, flush the rest down the proverbial toilet.

## Building the project

### Build requirements

- a Golang installation (built & tested on go v1.25)
- an internet connection to download dependencies (only necessary if dependencies have changed or this is the first build)
- a `make` installation. This project is built with GNU make v4 or higher; full compatibility with other versions of make (such as that shipped by Apple) is not guaranteed, but it _should_ be broadly compatible.

To build the project, simply run `make build` in the project's root directory to build the output executable.

> _Note: running with `make` is not strictly necessary. Reference the provided `Makefile` for typical development commands._
