# [Mojo](https://github.com/domokit/mojo) + [v23 RPCs](https://github.com/vanadium/docs/blob/master/concepts/rpc.md)

This repository implements the proposal outlined
[here](https://docs.google.com/a/google.com/document/d/17cMUkwolbQphimAYdyVNBCzA_f-HZy3YcxEpKOAmw48/edit?usp=sharing)
that enables communication between Mojo applications on different devices.

## Prerequisites

You must have the `jiri` tool installed with the `base` and `mojo` v23-profiles.

To update v23proxy to the latest version of mojo, you will need to also have
the `mojodev` profile.

## Quick start

You must always `make build` first. (The Makefile is not very good currently.)
- For desktop: `make build`
- For android: `ANDROID=1 make build`

The commands above build the `.mojo` shared library that can be run by mojo shells.
For example:
- `make start-v23proxy`
- `HOME=/tmp make ARGS="{see Makefile}" start-echo-client`

You can also run these with Android devices. Use an `ANDROID={N}` prefix to run on
the `Nth` Android device connected to your machine. `N` must be a positive integer.

Note: To run these examples, the devices used must run mojo_shell on the same local network.

## Updating v23proxy to the latest version of mojo

Prefix all commands with `USE_MOJO_DEV_PROFILE=1` in order to run with the
`mojodev` profile instead of `mojo`.
