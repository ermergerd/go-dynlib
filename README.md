Overview
========

go-dynlibs is a utility that attempts to compile the golang standard library packages as individual shared libraries.
The tool allows for passing ldflags and adopts the environment of the caller, allowing for easy cross compiling.
Currently, the utility is only testing to work with version go1.7 of the golang toolchain and standard libraries.
