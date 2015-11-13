PWD := $(shell pwd)

include ../shared/mojo.mk

# ANDROID needs to be any positive integer (e.g., 1, 2, 3, 4...)
ifdef ANDROID
	BUILD_DIR := $(PWD)/gen/mojo/android
	MOJO_SHARED_LIB := $(PWD)/gen/lib/android/libsystem_thunk.a

  # In order to determine the device id to target, we will parse the output of
  # `adb devices`. The target id is present within the ANDROID+1'th line.
	ANDROID_PLUS_ONE := $(shell echo $(ANDROID) \+ 1 | bc)
	DEVICE_ID := $(shell adb devices | sed -n $(ANDROID_PLUS_ONE)p | awk '{ print $$1; }')
	DEVICE_FLAG := --target-device $(DEVICE_ID)
	ANDROID_FLAG := --android
else
	BUILD_DIR := $(PWD)/gen/mojo/linux_amd64
	MOJO_SHARED_LIB := $(PWD)/gen/lib/linux_amd64/libsystem_thunk.a
endif

# If this is not the first mojo shell, then you must reuse the devservers
# to avoid a "port in use" error.
ifneq ($(shell fuser 31841/tcp),)
	REUSE_FLAG := --reuse-servers
endif

# Build the v23proxy and the associated examples.
.PHONY: build
build: $(BUILD_DIR)/v23proxy.mojo build-examples

build-examples: $(BUILD_DIR)/echo_client.mojo $(BUILD_DIR)/echo_server.mojo $(BUILD_DIR)/fortune_client.mojo $(BUILD_DIR)/fortune_server.mojo

# Go-based unit tests
test: gen/go/src/mojom/tests/transcoder_testcases/transcoder_testcases.mojom.go
	$(call MOGO_TEST,v.io/x/mojo/transcoder/...)

gen/go/src/mojom/tests/transcoder_testcases/transcoder_testcases.mojom.go: mojom/mojom/tests/transcoder_testcases.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

$(BUILD_DIR)/echo_client.mojo: $(MOJO_SHARED_LIB) gen/go/src/mojom/examples/echo/echo.mojom.go
	$(call MOGO_BUILD,examples/echo/client,$@)

$(BUILD_DIR)/echo_server.mojo: $(MOJO_SHARED_LIB) gen/go/src/mojom/examples/echo/echo.mojom.go
	$(call MOGO_BUILD,examples/echo/server,$@)

gen/go/src/mojom/examples/echo/echo.mojom.go: mojom/mojom/examples/echo.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

$(BUILD_DIR)/fortune_client.mojo: $(MOJO_SHARED_LIB) gen/go/src/mojom/examples/fortune/fortune.mojom.go
	$(call MOGO_BUILD,examples/fortune/client,$@)

$(BUILD_DIR)/fortune_server.mojo: $(MOJO_SHARED_LIB) gen/go/src/mojom/examples/fortune/fortune.mojom.go
	$(call MOGO_BUILD,examples/fortune/server,$@)

gen/go/src/mojom/examples/fortune/fortune.mojom.go: mojom/mojom/examples/fortune.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

$(BUILD_DIR)/v23proxy.mojo: $(MOJO_SHARED_LIB) gen/go/src/mojom/v23proxy/v23proxy.mojom.go | mojo-env-check
	$(call MOGO_BUILD,v.io/x/mojo/proxy,$@)

mojom/mojo/public/interfaces/bindings/mojom_types.mojom: $(MOJO_DIR)/src/mojo/public/interfaces/bindings/mojom_types.mojom
	mkdir -p mojom/mojo/public/interfaces/bindings
	ln -sf $(MOJO_DIR)/src/mojo/public/interfaces/bindings/mojom_types.mojom mojom/mojo/public/interfaces/bindings/mojom_types.mojom

gen/go/src/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.go: mojom/mojo/public/interfaces/bindings/mojom_types.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

gen/go/src/mojom/v23proxy/v23proxy.mojom.go: mojom/mojom/v23proxy.mojom gen/go/src/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.go | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

# Run the Mojo shell with map-origin. This is common to Linux and Android since
# the latter cannot accept a config-file.
define RUN_MOJO_SHELL
	$(MOJO_DIR)/src/mojo/devtools/common/mojo_run \
	$1 \
	$(ANDROID_FLAG) \
	$(DEVICE_FLAG) \
	--no-config-file \
	$(REUSE_FLAG) \
	--map-origin="https://mojo.v.io/=$(BUILD_DIR)" \
	--map-origin="https://mojo.v.io/=." \
	--args-for="https://mojo.v.io/$2 $3" \
	https://mojo.v.io/$2
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
start-v23proxy: build
	$(call RUN_MOJO_SHELL,--enable-multiprocess,v23proxy.mojo,)

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
start-echo-client: build
	$(call RUN_MOJO_SHELL,,echo_client.mojo,${ARGS})

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
start-fortune-client: build
	$(call RUN_MOJO_SHELL,,fortune_client.mojo,${ARGS})

.PHONY: clean
clean:
	rm -r gen
