# systemd [![Build Status](https://travis-ci.org/julienschmidt/systemd.svg?branch=master)](https://travis-ci.org/julienschmidt/systemd) [![Coverage Status](https://coveralls.io/repos/github/julienschmidt/systemd/badge.svg?branch=master)](https://coveralls.io/github/julienschmidt/systemd?branch=master) [![GoDoc](https://godoc.org/github.com/julienschmidt/systemd?status.svg)](https://godoc.org/github.com/julienschmidt/systemd)

This package provides native systemd integration for Go programs.

## Socket Activation

systemd socket activation (or any other compatible socket passing system passing sockets via `LISTEN_FDS` enviornment variables) is enabled by [systemd.Listen](https://godoc.org/github.com/julienschmidt/systemd#Listen) and [systemd.ListenWithNames](https://godoc.org/github.com/julienschmidt/systemd#ListenWithNames).
