# [Mojo](https://github.com/domokit/mojo) + [v23 RPCs](https://github.com/vanadium/docs/blob/master/concepts/rpc.md)

This repository implements the proposal outlined
[here](https://docs.google.com/a/google.com/document/d/17cMUkwolbQphimAYdyVNBCzA_f-HZy3YcxEpKOAmw48/edit?usp=sharing)
that enables communication between Mojo applications on different devices.

## Quick start

- For desktop: `make build`
- For android: `ANDROID=1 make build`

The commands above build the `.mojo` shared library that can be run by mojo shells.
For example:
- `make start-v23proxy`
- `REMOTE_ENDPOINT=<see Makefile> HOME=/tmp make start-echo-client`
