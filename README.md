# systemd [![GoDoc](https://godoc.org/github.com/julienschmidt/systemd?status.svg)](https://godoc.org/github.com/julienschmidt/systemd)

This package provides native systemd integration for Go programs.

## Socket Activation

systemd socket activation is enabled by [systemd.Listen](https://godoc.org/github.com/julienschmidt/systemd#Listen) and [systemd.ListenWithNames](https://godoc.org/github.com/julienschmidt/systemd#ListenWithNames).
