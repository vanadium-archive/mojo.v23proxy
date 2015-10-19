PWD := $(shell pwd)

include ../shared/mojo.mk

ifdef ANDROID
	BUILD_DIR := $(PWD)/gen/mojo/android
	MOJO_SHARED_LIB := $(PWD)/gen/lib/android/libsystem_thunk.a
else
	BUILD_DIR := $(PWD)/gen/mojo/linux_amd64
	MOJO_SHARED_LIB := $(PWD)/gen/lib/linux_amd64/libsystem_thunk.a
endif

# Build the v23proxy and the associated examples.
.PHONY: build
build: $(BUILD_DIR)/v23proxy.mojo build-examples

build-examples: $(BUILD_DIR)/echo_client.mojo $(BUILD_DIR)/echo_server.mojo $(BUILD_DIR)/fortune_client.mojo $(BUILD_DIR)/fortune_server.mojo

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

define RUN_MOJO_SHELL
	$(MOJO_DIR)/src/mojo/devtools/common/mojo_run \
	$1 \
	--config-file mojoconfig \
	--config-alias BUILD_DIR=$(BUILD_DIR) \
	--config-alias PORT=$2 \
	https://mojo.v.io/$3
endef

# Start the v23proxy (server-side). This runs the v23proxy in its own shell and
# will print an endpoint to stdout. That endpoint needs to be passed to the clients.
#
# Highly recommended: Prefix this command with HOME={a unique tmp directory}
# We cannot disable the db LOCK that is created, so one workaround is to use a
# different HOME directory per mojo shell.
.PHONY: start-v23proxy
start-v23proxy: build
	$(call RUN_MOJO_SHELL,--enable-multiprocess,31941,v23proxy.mojo)

# Start the echo client. This uses the v23proxy (client-side) to speak Vanadium
# over to the v23proxy (server-side) [OR a 0-authentication Vanadium echo server].
#
# Note: Does not use --enable-multiprocess since small Go programs can omit it.
# Don't forget to prepend REMOTE_ENDPOINT={endpoint}/https://mojo.v.io/echo_server.mojo/mojo::examples::RemoteEcho
#
# Highly recommended: Prefix this command with HOME={a unique tmp directory}
# We cannot disable the db LOCK that is created, so one workaround is to use a
# different HOME directory per mojo shell.
.PHONY: start-echo-client
start-echo-client: build
	$(call RUN_MOJO_SHELL,,31942,echo_client.mojo)

# Start the fortune client. This uses the v23proxy (client-side) to speak Vanadium
# over to the v23proxy (server-side) [OR a 0-authentication Vanadium fortune server].
#
# Note: Does not use --enable-multiprocess since small Go programs can omit it.
# Don't forget to prepend REMOTE_ENDPOINT={endpoint}/https://mojo.v.io/fortune_server.mojo/mojo::examples::Fortune
# You may optionally prepend ADD_FORTUNE={some fortune}
#
# Highly recommended: Prefix this command with HOME={a unique tmp directory}
# We cannot disable the db LOCK that is created, so one workaround is to use a
# different HOME directory per mojo shell.
.PHONY: start-fortune-client
start-fortune-client: build
	$(call RUN_MOJO_SHELL,,31943,fortune_client.mojo)

.PHONY: clean
clean:
	rm -r gen
