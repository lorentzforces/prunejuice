# prunejuice

*Violently ejecting the refuse*

## Building the project

### Build requirements

- a Golang installation (built & tested on go v1.25)
- an internet connection to download dependencies (only necessary if dependencies have changed or this is the first build)
- a `make` installation. This project is built with GNU make v4 or higher; full compatibility with other versions of make (such as that shipped by Apple) is not guaranteed, but it _should_ be broadly compatible.

To build the project, simply run `make build` in the project's root directory to build the output executable.

> _Note: running with `make` is not strictly necessary. Reference the provided `Makefile` for typical development commands._
