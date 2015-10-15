PWD := $(shell pwd)

include ../shared/mojo.mk

ifdef ANDROID
	BUILD_DIR := $(PWD)/gen/mojo/android
	MOJO_SHARED_LIB := $(PWD)/gen/lib/android/libsystem_thunk.a
else
	BUILD_DIR := $(PWD)/gen/mojo/linux_amd64
	MOJO_SHARED_LIB := $(PWD)/gen/lib/linux_amd64/libsystem_thunk.a
endif

build: $(BUILD_DIR)/v23proxy.mojo

$(BUILD_DIR)/v23proxy.mojo: $(MOJO_SHARED_LIB) gen/go/src/mojom/v23proxy/v23proxy.mojom.go | mojo-env-check
	$(call MOGO_BUILD,v.io/x/mojo/proxy,$@)

mojom/mojo/public/interfaces/bindings/mojom_types.mojom: $(MOJO_DIR)/src/mojo/public/interfaces/bindings/mojom_types.mojom
	mkdir -p mojom/mojo/public/interfaces/bindings
	ln -s $(MOJO_DIR)/src/mojo/public/interfaces/bindings/mojom_types.mojom mojom/mojo/public/interfaces/bindings/mojom_types.mojom

gen/go/src/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.go: mojom/mojo/public/interfaces/bindings/mojom_types.mojom | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@

gen/go/src/mojom/v23proxy/v23proxy.mojom.go: mojom/mojom/v23proxy.mojom gen/go/src/mojo/public/interfaces/bindings/mojom_types/mojom_types.mojom.go | mojo-env-check
	$(call MOJOM_GEN,$<,mojom,gen,go)
	gofmt -w $@


.PHONY: build
