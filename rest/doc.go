// Copyright (c) 2014 Datacratic. All rights reserved.

/*

Package rest provides the ability to construct JSON REST endpoints and clients.

Endpoints are provided by constructing an Endpoint struct and registering Route
objects. Route objects are constructed from handlers and Path objects. Each HTTP
request received by the Endpoint will be routed to the handler of the route with
a matching path.

Paths can contain variable components denoted by a leading ':' character. These
variable arguments will be used to automatically populate the arguments of the
handler. Constant components take precedence over variable components during
routing. Note that a request can only be routed to a single handler and
duplicate paths are therefore rejected.

Clients are provided by the Client struct which allows the incremental
construction of REST request. The response is sent when calling the
Client.Send() function and the return can be processed via the
Response.GetBody() function which will check for various HTTP error conditions.

gorest currently only supports JSON requests with the exception of error
messages which are communicated as strings. Request and response body must
therefore be compatible with the encoding/json package.

*/
package rest
