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

# Installs dart dependencies.
packages:
	pub get

.PHONY: upgrade-packages
upgrade-packages:
	pub upgrade

build-go-examples: $(BUILD_DIR)/echo_client.mojo $(BUILD_DIR)/echo_server.mojo $(BUILD_DIR)/fortune_client.mojo $(BUILD_DIR)/fortune_server.mojo

build-dart-examples: gen/echo.mojom.dart gen/fortune.mojom.dart

# Go-based unit tests
test: $(MOJO_SHARED_LIB) gen/go/src/mojom/tests/transcoder_testcases/transcoder_testcases.mojom.go
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

$(BUILD_DIR)/v23proxy.mojo: gen/go/src/mojom/v23proxy/v23proxy.mojom.go | mojo-env-check
	$(call MOGO_BUILD,v.io/x/mojo/proxy,$@)

gen/go/src/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.go: mojom/mojo/public/interfaces/bindings/mojom_types.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

gen/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.dart: mojom/mojo/public/interfaces/bindings/mojom_types.mojom packages | mojo-env-check
	$(call MOJOM_GEN,$<,.,lib/gen,dart)
	# TODO(nlacasse): mojom_bindings_generator creates bad symlinks on dart
	# files, so we delete them.  Stop doing this once the generator is fixed.
	# See https://github.com/domokit/mojo/issues/386
	rm -f lib/gen/mojom/$(notdir $@)

gen/go/src/mojom/v23proxy/v23proxy.mojom.go: | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

gen/v23proxy.mojom.dart: mojom/mojom/v23proxy.mojom packages gen/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.dart | mojo-env-check
	$(call MOJOM_GEN,$<,.,lib/gen,dart)
	# TODO(nlacasse): mojom_bindings_generator creates bad symlinks on dart
	# files, so we delete them.  Stop doing this once the generator is fixed.
	# See https://github.com/domokit/mojo/issues/386
	rm -f lib/gen/mojom/$(notdir $@)

# Run the Mojo shell with map-origin. This is common to Linux and Android since
# the latter cannot accept a config-file.
# $1 is for any extra flags, like --enable-multiprocess.
# $2 is for the name and/or path to the mojo or dart file.
# $3 is for $ARGS, any arguments you might want to pass to the mojo program.
# $4 is for 'dart'. This is temporary so that the --map-origin works out.
# TODO(alexfandrianto): Figure out how to make this mapping work without
# needing a distinct URL.
define RUN_MOJO_SHELL
	$(MOJO_DEVTOOLS)/mojo_run \
	$1 \
	$(ANDROID_FLAG) \
	$(DEVICE_FLAG) \
	--no-config-file \
	$(REUSE_FLAG) \
	$(ORIGIN_FLAG) \
	--map-origin="https://mojo.v.io/=$(BUILD_DIR)" \
	--map-origin="https://mojodart.v.io/=$(PWD)" \
	--args-for="https://mojo$4.v.io/$2 $3" \
	https://mojo$4.v.io/$2
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
	$(call RUN_MOJO_SHELL,--enable-multiprocess,v23proxy.mojo,,)


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
	$(call RUN_MOJO_SHELL,,echo_client.mojo,${ARGS},)

# Like the start-echo-client but using a Dart client instead.
# Note: Uses --enable-multiprocess since it looks like the Dart VM and Go VM
# together are enough to cause a SIGSEGV (Android signal 11 crash) if this flag
# is not used.
.PHONY: start-dart-echo-client
start-dart-echo-client: build-dart
	$(call RUN_MOJO_SHELL,--enable-multiprocess,dart-examples/echo/lib/main.dart,${ARGS},dart)


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
	$(call RUN_MOJO_SHELL,,fortune_client.mojo,${ARGS},)

# Like the start-fortune-client but using a Dart client instead.
# Note: Uses --enable-multiprocess since it looks like the Dart VM and Go VM
# together are enough to cause a SIGSEGV (Android signal 11 crash) if this flag
# is not used.
.PHONY: start-dart-fortune-client
start-dart-fortune-client: build-dart
	$(call RUN_MOJO_SHELL,--enable-multiprocess,dart-examples/fortune/lib/main.dart,${ARGS},dart)


.PHONY: clean
clean: clean-go clean-dart

.PHONY: clean-go
clean-go:
	rm -rf gen

.PHONY: clean-dart
clean-dart:
	rm -rf lib/gen
	rm -rf dart-examples/echo/lib/gen
