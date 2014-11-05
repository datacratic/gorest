# gorest #

Yet another JSON REST library for golang.


## Installation ##

You can download the code via the usual go utilities:

```
go get github.com/datacratic/gorest/rest
```

To build the code and run the test suite along with several static analysis
tools, use the provided Makefile:

```
make test
```

Note that the usual go utilities will work just fine but we require that all
commits pass the full suite of tests and static analysis tools.


## Examples ##

Usage examples are available in the following
[test suite](rest/example_test.go).


## Why Another REST Library? ##

This library is intended to be used in low-latency scenarios where we need to a
tighter control over the allocations and the complexity of the internal
data-structures while still providing a dirt simple interface.

gorest will also eventually support a documentation endpoint which conforms the
the internal datcratic REST endpoint documentation format. This will be used to
implement an interactive web-interface to the REST endpoints.


## License ##

The source code is available under the Apache License. See the LICENSE file for
more details.
