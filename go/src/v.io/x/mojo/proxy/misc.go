// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Construct the proper *vdl.Value (as a struct) from the mojom type.
import (
	"fmt"

	"mojo/public/go/bindings"

	"v.io/v23/vdl"
	"v.io/x/mojo/transcoder"
)

// TODO(alexfandrianto): Since this function could panic, we should consider
// handling that if it happens.
func combineVdlValueByMojomType(values []*vdl.Value, t *vdl.Type) *vdl.Value {
	outVdlValue := vdl.ZeroValue(t)
	for i := 0; i < t.NumField(); i++ {
		outVdlValue.StructField(i).Assign(values[i])
	}
	return outVdlValue
}

// Construct []*vdl.Value from a *vdl.Value (as a struct) via its mojom type.
// TODO(alexfandrianto): Since this function could panic, we should consider
// handling that if it happens.
func splitVdlValueByMojomType(value *vdl.Value, t *vdl.Type) []*vdl.Value {
	outVdlValues := make([]*vdl.Value, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		outVdlValues[i] = value.StructField(i)
	}
	return outVdlValues
}

func encodeMessageFromVom(header bindings.MessageHeader, argptrs []interface{}, t *vdl.Type) (*bindings.Message, error) {
	// Convert argptrs into their true form: []*vdl.Value
	inargs := make([]*vdl.Value, len(argptrs))
	for i := range argptrs {
		inargs[i] = *argptrs[i].(**vdl.Value)
	}

	// Construct the proper *vdl.Value (as a struct) from the mojom type.
	vdlValue := combineVdlValueByMojomType(inargs, t)

	encoder := bindings.NewEncoder()
	if err := header.Encode(encoder); err != nil {
		return nil, err
	}
	if bytes, handles, err := encoder.Data(); err != nil {
		return nil, err
	} else {
		// Encode here.
		moreBytes, err := transcoder.VdlToMojom(vdlValue)
		if err != nil {
			return nil, fmt.Errorf("mojovdl.Encode failed: %v", err)
		}
		// Append the encoded "payload" to the end of the slice.
		bytes = append(bytes, moreBytes...)

		return &bindings.Message{
			Header:  header,
			Bytes:   bytes,
			Handles: handles,
			Payload: moreBytes,
		}, nil
	}
}
