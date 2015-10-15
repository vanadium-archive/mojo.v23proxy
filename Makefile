PWD := $(shell pwd)
V23_GO_FILES := $(shell find $(JIRI_ROOT) -name "*.go")

include ../shared/mojo.mk

# Flags for V23Proxy mojo service.
V23_MOJO_FLAGS := --v=0

build:
	$(call MOGO_BUILD,v.io/x/mojo/proxy,$@)
