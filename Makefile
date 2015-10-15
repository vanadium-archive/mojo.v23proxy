PWD := $(shell pwd)

include ../shared/mojo.mk

ifdef ANDROID
	BUILD_DIR := $(PWD)/gen/mojo/android
	MOJO_SHARED_LIB := $(PWD)/gen/lib/android/libsystem_thunk.a
else
	BUILD_DIR := $(PWD)/gen/mojo/linux_amd64
	MOJO_SHARED_LIB := $(PWD)/gen/lib/linux_amd64/libsystem_thunk.a
endif

$(BUILD_DIR)/v23proxy.mojo: $(MOJO_SHARED_LIB)
	$(call MOGO_BUILD,v.io/x/mojo/proxy,$@)
