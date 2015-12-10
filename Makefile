PWD := $(shell pwd)

include ../shared/mojo.mk

# ANDROID needs to be any positive integer (e.g., 1, 2, 3, 4...)
ifdef ANDROID
	BUILD_DIR := $(PWD)/gen/mojo/android

	# For some reason we need to set the origin flag when running on Android,
	# but setting it on Linux causes errors.
	ORIGIN_FLAG = --origin $(MOJO_SERVICES)

	# In order to determine the device id to target, we will parse the output
	# of `adb devices`. The target id is present within the ANDROID+1'th line.
	ANDROID_PLUS_ONE := $(shell echo $(ANDROID) \+ 1 | bc)
	DEVICE_ID := $(shell adb devices | sed -n $(ANDROID_PLUS_ONE)p | awk '{ print $$1; }')
	DEVICE_FLAG := --target-device $(DEVICE_ID)
	ANDROID_FLAG := --android
else
	BUILD_DIR := $(PWD)/gen/mojo/linux_amd64
endif

# If this is not the first mojo shell, then you must reuse the devservers
# to avoid a "port in use" error.
ifneq ($(shell fuser 31841/tcp),)
	REUSE_FLAG := --reuse-servers
endif

# Build the v23proxy and the associated examples in Go and Dart.
.PHONY: build
build: build-go build-dart

# Build the v23proxy.mojo, client, and associated go examples.
.PHONY: build-go
build-go: $(BUILD_DIR)/v23proxy.mojo build-go-examples

# Build the v23proxy client and its associated dart examples.
.PHONY: build-dart
build-dart: gen/v23proxy.mojom.dart build-dart-examples lint-dart

.PHONY: lint-dart
lint-dart:
	dartanalyzer lib/client.dart | grep -v "\[warning\] The imported libraries"
	dartanalyzer dart-examples/echo/lib/main.dart | grep -v "\[warning\] The imported libraries"

.PHONY: link-mojo_sdk
link-mojo_sdk:
	ln -sf $(MOJO_SDK) .mojo_sdk

# Installs dart dependencies.
packages: link-mojo_sdk
	pub get

.PHONY: upgrade-packages
upgrade-packages: link-mojo_sdk
	pub upgrade

build-go-examples: $(BUILD_DIR)/echo_client.mojo $(BUILD_DIR)/echo_server.mojo $(BUILD_DIR)/fortune_client.mojo $(BUILD_DIR)/fortune_server.mojo

build-dart-examples: gen/echo.mojom.dart gen/fortune.mojom.dart

.PHONY: test
test: test-unit test-integration

# Go-based unit tests
.PHONY: test-unit
test-unit: $(MOJO_SHARED_LIB) gen/go/src/mojom/tests/transcoder_testcases/transcoder_testcases.mojom.go gen-vdl
	$(call MOGO_TEST,v.io/x/mojo/transcoder/...)

# Note:This file is needed to compile v23proxy.mojom, so we're symlinking it in from $MOJO_SDK.
mojom/mojo/public/interfaces/bindings/mojom_types.mojom: $(MOJO_SDK)/src/mojo/public/interfaces/bindings/mojom_types.mojom
	mkdir -p mojom/mojo/public/interfaces/bindings
	ln -sf $(MOJO_SDK)/src/mojo/public/interfaces/bindings/mojom_types.mojom mojom/mojo/public/interfaces/bindings/mojom_types.mojom

mojom/mojo/public/interfaces/bindings/tests/test_unions.mojom: $(MOJO_SDK)/src/mojo/public/interfaces/bindings/tests/test_unions.mojom
	mkdir -p mojom/mojo/public/interfaces/bindings/tests
	ln -sf $(MOJO_SDK)/src/mojo/public/interfaces/bindings/tests/test_unions.mojom mojom/mojo/public/interfaces/bindings/tests/test_unions.mojom

mojom/mojo/public/interfaces/bindings/tests/test_included_unions.mojom: $(MOJO_SDK)/src/mojo/public/interfaces/bindings/tests/test_included_unions.mojom
	mkdir -p mojom/mojo/public/interfaces/bindings/tests
	ln -sf $(MOJO_SDK)/src/mojo/public/interfaces/bindings/tests/test_included_unions.mojom mojom/mojo/public/interfaces/bindings/tests/test_included_unions.mojom

mojom/mojo/public/interfaces/bindings/tests/test_structs.mojom: $(MOJO_SDK)/src/mojo/public/interfaces/bindings/tests/test_structs.mojom
	mkdir -p mojom/mojo/public/interfaces/bindings/tests
	ln -sf $(MOJO_SDK)/src/mojo/public/interfaces/bindings/tests/test_structs.mojom mojom/mojo/public/interfaces/bindings/tests/test_structs.mojom

gen/go/src/mojom/tests/transcoder_testcases/transcoder_testcases.mojom.go: mojom/mojom/tests/transcoder_testcases.mojom mojom/mojo/public/interfaces/bindings/tests/test_unions.mojom mojom/mojo/public/interfaces/bindings/tests/test_included_unions.mojom mojom/mojo/public/interfaces/bindings/tests/test_structs.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

$(BUILD_DIR)/echo_client.mojo: gen/go/src/mojom/examples/echo/echo.mojom.go
	$(call MOGO_BUILD,examples/echo/client,$@)

$(BUILD_DIR)/echo_server.mojo: gen/go/src/mojom/examples/echo/echo.mojom.go
	$(call MOGO_BUILD,examples/echo/server,$@)

.PHONY: test-integration
test-integration: $(BUILD_DIR)/test_client.mojo $(BUILD_DIR)/test_server.mojo $(BUILD_DIR)/v23proxy.mojo
	GOPATH=$(PWD)/go:$(PWD)/gen/go jiri go -profiles=base,$(MOJO_PROFILE) run go/src/v.io/x/mojo/tests/cmd/runtest.go

$(BUILD_DIR)/test_client.mojo: go/src/v.io/x/mojo/tests/client/test_client.go gen/go/src/mojom/tests/end_to_end_test/end_to_end_test.mojom.go gen/go/src/mojom/v23proxy/v23proxy.mojom.go
	$(call MOGO_BUILD,v.io/x/mojo/tests/client,$@)

$(BUILD_DIR)/test_server.mojo: go/src/v.io/x/mojo/tests/server/test_server.go gen/go/src/mojom/tests/end_to_end_test/end_to_end_test.mojom.go
	$(call MOGO_BUILD,v.io/x/mojo/tests/server,$@)

gen/go/src/mojom/examples/echo/echo.mojom.go: mojom/mojom/examples/echo.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

gen/echo.mojom.dart: mojom/mojom/examples/echo.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,dart-examples/echo/lib/gen,dart)

$(BUILD_DIR)/fortune_client.mojo: gen/go/src/mojom/examples/fortune/fortune.mojom.go
	$(call MOGO_BUILD,examples/fortune/client,$@)

$(BUILD_DIR)/fortune_server.mojo: gen/go/src/mojom/examples/fortune/fortune.mojom.go
	$(call MOGO_BUILD,examples/fortune/server,$@)

gen/go/src/mojom/examples/fortune/fortune.mojom.go: mojom/mojom/examples/fortune.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

gen/fortune.mojom.dart: mojom/mojom/examples/fortune.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,dart-examples/fortune/lib/gen,dart)

$(BUILD_DIR)/v23proxy.mojo: $(shell find $(PWD)/go/src/v.io/x/mojo/proxy -name *.go) | mojo-env-check
	$(call MOGO_BUILD,v.io/x/mojo/proxy,$@)

gen/go/src/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.go: mojom/mojo/public/interfaces/bindings/mojom_types.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

gen/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.dart: mojom/mojo/public/interfaces/bindings/mojom_types.mojom packages | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,lib/gen,dart)
	# TODO(nlacasse): mojom_bindings_generator creates bad symlinks on dart
	# files, so we delete them.  Stop doing this once the generator is fixed.
	# See https://github.com/domokit/mojo/issues/386
	rm -f lib/gen/mojom/$(notdir $@)

gen/go/src/mojom/v23proxy/v23proxy.mojom.go: mojom/mojom/v23proxy.mojom mojom/mojo/public/interfaces/bindings/mojom_types.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

gen/go/src/mojom/tests/end_to_end_test/end_to_end_test.mojom.go: mojom/mojom/tests/end_to_end_test.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

gen/v23proxy.mojom.dart: mojom/mojom/v23proxy.mojom packages gen/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.dart | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,lib/gen,dart)
	# TODO(nlacasse): mojom_bindings_generator creates bad symlinks on dart
	# files, so we delete them.  Stop doing this once the generator is fixed.
	# See https://github.com/domokit/mojo/issues/386
	rm -f lib/gen/mojom/$(notdir $@)

gen-vdl:
	GOPATH=$(PWD)/go VDLPATH=$(PWD)/go vdl generate all

# Run the Mojo shell with map-origin. This is common to Linux and Android since
# the latter cannot accept a config-file.
# $1 is for the name and/or path to the mojo or dart file.
# $2 is for $ARGS, any arguments you might want to pass to the mojo program.
# TODO(alexfandrianto): Figure out how to make this mapping work without
# needing a distinct URL.
define RUN_MOJO_SHELL
	$(MOJO_DEVTOOLS)/mojo_run \
	https://mojo.v.io/$1 \
	--config-file $(PWD)/mojoconfig \
	--shell-path $(MOJO_SHELL) \
	$(ANDROID_FLAG) \
  	--enable-multiprocess \
  	--config-alias V23PROXY_DIR=$(PWD) \
  	--config-alias V23PROXY_BUILD_DIR=$(BUILD_DIR) \
  	"--args-for=https://mojo.v.io/$1 $(ARGS) $(V23_MOJO_FLAGS)" \
  	"--args-for=mojo:dart_content_handler --enable-strict-mode" \
  	$(ORIGIN_FLAG)
	$(MOJO_SHELL_FLAGS)
endef

# Start the v23proxy (server-side). This runs the v23proxy in its own shell and
# will print an endpoint to stdout. That endpoint needs to be passed to the clients.
#
# On Linux, run with
# make start-v23proxy
# (Optionally, this can be prefixed with a HOME directory.)
#
# On Android, run with
# ANDROID={device number} make start-v23proxy
.PHONY: start-v23proxy
start-v23proxy: build-go
	$(call RUN_MOJO_SHELL,v23proxy.mojo,)


# Start the echo client. This uses the v23proxy (client-side) to speak Vanadium
# over to the v23proxy (server-side) [OR a 0-authentication Vanadium echo server].
#
# On Linux, run with
# HOME={tmpdir} make ARGS="{remote endpoint}//https://mojo.v.io/echo_server.mojo/mojo::examples::RemoteEcho [optional: a string to echo]" start-echo-client
#
# On Android, run with
# ANDROID={device number} make ARGS="{remote endpoint}//https://mojo.v.io/echo_server.mojo/mojo::examples::RemoteEcho [optional: a string to echo]" start-echo-client
#
# Note1: Does not use --enable-multiprocess since small Go programs can omit it.
# Note2: Setting HOME ensures that we avoid a db LOCK that is created per mojo shell instance.
.PHONY: start-echo-client
start-echo-client: build-go
	$(call RUN_MOJO_SHELL,echo_client.mojo,${ARGS})

# Like the start-echo-client but using a Dart client instead.
# Note: Uses --enable-multiprocess since it looks like the Dart VM and Go VM
# together are enough to cause a SIGSEGV (Android signal 11 crash) if this flag
# is not used.
.PHONY: start-dart-echo-client
start-dart-echo-client: build-dart
	$(call RUN_MOJO_SHELL,dart-examples/echo/lib/main.dart,${ARGS})


# Start the fortune client. This uses the v23proxy (client-side) to speak Vanadium
# over to the v23proxy (server-side) [OR a 0-authentication Vanadium fortune server].
#
# On Linux, run with
# HOME={tmpdir} make ARGS="{remote endpoint}//https://mojo.v.io/fortune_server.mojo/mojo::examples::Fortune [optional: a fortune to add]" start-fortune-client
#
# On Android, run with
# ANDROID={device number} make ARGS="{remote endpoint}//https://mojo.v.io/fortune_server.mojo/mojo::examples::Fortune [optional: a fortune to add]" start-fortune-client
#
# Note1: Does not use --enable-multiprocess since small Go programs can omit it.
# Note2: Setting HOME ensures that we avoid a db LOCK that is created per mojo shell instance.
.PHONY: start-fortune-client
start-fortune-client: build-go
	$(call RUN_MOJO_SHELL,fortune_client.mojo,${ARGS})

# Like the start-fortune-client but using a Dart client instead.
# Note: Uses --enable-multiprocess since it looks like the Dart VM and Go VM
# together are enough to cause a SIGSEGV (Android signal 11 crash) if this flag
# is not used.
.PHONY: start-dart-fortune-client
start-dart-fortune-client: build-dart
	$(call RUN_MOJO_SHELL,dart-examples/fortune/lib/main.dart,${ARGS})


.PHONY: clean
clean: clean-go clean-dart

.PHONY: clean-go
clean-go:
	rm -rf gen
	rm -rf mojom/mojo

.PHONY: clean-dart
clean-dart:
	rm -rf lib/gen
	rm -rf dart-examples/echo/lib/gen
	rm -rf dart-examples/fortune/lib/gen
	rm -rf packages
