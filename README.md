# logstreamer [![Build Status](https://travis-ci.org/mfycheng/logstreamer.svg?branch=master)](https://travis-ci.org/mfycheng/logstreamer) [![GoDoc](https://godoc.org/github.com/mfycheng/logstreamer?status.svg)](https://godoc.org/github.com/mfycheng/logstreamer)

logstreamer is a tool that allows the streaming of logs to an arbitrary amount of users.

## Motivation

Sometimes, when working on project with other people, not everyone has access to your server machine(s).
It becomes annoying to have to SSH into the machine and show / provide them with the log outputs for debugging
purposes. Sometimes it's annoying to have to SSH into the machine for logs in general.

logstreamer is a quick and dirty solution to this. Given an input stream, it allows multiple observers
to 'observe' the stream (either using websockets, cURL, etc.). Additionally, logstreamer can remember the
last X amount of log lines, so if something bad happens, an observer will hopefully be able to see it
restroactively.

In some cases, your development machine may be a tiny instance (t2.micro, smallest Digital Ocean droplet),
with very limited disk space. Letting logs persist to disk can often fill up disk space quickly. logstreamer
can help with this by just receiving the logs of stdin. Since it's a development machine, logs generally
don't matter unless something goes wrong, or someone is watching.

## Building

```
go get github.com/mfycheng/logstreamer
```

## Using

logstreamer comes with a tool `logstream` that offers multiple options for streaming logs. Currently,
it only offers websocket access. An example use of `logstream` is:

```
application | go run cmd/logstream/main.go
```

or, if installed to your `$PATH` / `$GOPATH`:

```
application | logstream
```

## TODO

There are still things to do, such as:

 * Configurable options for the server
 * Output log to file (with possible rotation)
