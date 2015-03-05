# quickbump-quiz-webapp

This is a short readme for the QuickBump project. It describes usage as well as
building instructions.


Usage:

The server has a few options. You can run the QuickBump server with -h or -help.

It must serve the QuickBump HTTP API (see -apiurl), but can optionally serve
static files -- the QuickBump client by default (see -wwwroot).

The server can be configured to listen on whatever address and port with the
-addr option.

The server can be built with different database modules and one is selected at
runtime (see -db).

The server can also be provided a dictionary for creating question identifiers
(see -words). But This feature is largely useless since we never got around to
implementing security.

If built with QR Code support, the server can render QR Codes to encode stuff
(fixme) (see -qrurl).


How to build QuickBump:

For the basic server system, the files you need are
    main.go answer.go data.go handler.go question.go
as well as at least one file that provides a database implementation. At the
moment, this is just memdb.go and mongodb.go
You can compile and run the QuickBump server with the `go run` command. For
example, if using the memdb.go file:
    go run main.go answer.go data.go handler.go memdb.go question.go
Or just build an executable named `main' with:
    go build main.go answer.go data.go handler.go memdb.go question.go
To compile with QR Code support, include qrcode.go in the above commands. This
requires the QREncode (libqrencode) library to be installed.

The client does not need any sort of building.