# gorest #

Yet another JSON REST library for golang. Usage examples are available in the
[example_test.go](rest/example_test.go) test suite.

This library is intended to be used in low-latency scenarios where we need to a
tighter control over the allocations and the complexity of the internal
data-structures while still providing a dirt simple interface.

gorest will also eventually support a documentation endpoint which conforms the
the internal datcratic REST endpoint documentation format. This will be used to
implement an interactive web-interface to the REST endpoints.
