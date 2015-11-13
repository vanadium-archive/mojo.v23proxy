// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder_test

import (
	"reflect"
	"testing"

	"mojo/public/go/bindings"

	"fmt"

	"bytes"

	"v.io/v23/vdl"
	"v.io/x/mojo/transcoder"
)

func TestMojoToVom(t *testing.T) {
	for _, test := range testCases {
		testName := test.Name + " mojo->vom"

		data, err := computeExpectedMojomBytes(test.MojoValue)
		if err != nil {
			t.Errorf("%s: %v", testName, err)
			continue
		}

		var out interface{}
		if err := transcoder.MojomToVdl(data, vdl.TypeOf(test.VdlValue), &out); err != nil {
			t.Errorf("%s: error in MojoToVom: %v (was transcoding from %x)", testName, err, data)
			continue
		}

		if got, want := out, test.VdlValue; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: result doesn't match expectation. got %#v, but want %#v", testName, got, want)
		}
	}
}

func TestVomToMojo(t *testing.T) {
	for _, test := range testCases {
		testName := test.Name + " vom->mojo"

		data, err := transcoder.VdlToMojom(test.VdlValue)
		if err != nil {
			t.Errorf("%s: error in VomToMojo: %v", testName, err)
			continue
		}

		expectedData, err := computeExpectedMojomBytes(test.MojoValue)
		if err != nil {
			t.Errorf("%s: %v", testName, err)
			continue
		}

		if got, want := data, expectedData; !bytes.Equal(got, want) {
			t.Errorf("%s: got %x, but want %x", testName, got, want)
		}
	}
}

func computeExpectedMojomBytes(mojoValue interface{}) ([]byte, error) {
	payload, ok := mojoValue.(bindings.Payload)
	if !ok {
		return nil, fmt.Errorf("type %T lacks an Encode() method", mojoValue)
	}

	enc := bindings.NewEncoder()
	err := payload.Encode(enc)
	if err != nil {
		return nil, fmt.Errorf("error in Encode: %v", err)
	}
	data, _, err := enc.Data()
	if err != nil {
		return nil, fmt.Errorf("error in Data()", err)
	}
	return data, nil
}
