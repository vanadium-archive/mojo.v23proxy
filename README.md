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
- `HOME=/tmp make ARGS="{see Makefile}" start-echo-client`

You can also run these with Android devices. Use an `ANDROID={N}` prefix to run on
the `Nth` Android device connected to your machine. `N` must be a positive integer.

Note: To run these examples, the devices used must run mojo_shell on the same local network.