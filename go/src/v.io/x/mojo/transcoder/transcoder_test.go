// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder_test

import (
	"fmt"
	"reflect"
	"testing"

	"mojo/public/go/bindings"

	"v.io/v23/vdl"
	"v.io/x/mojo/transcoder"
)

func TestMojoToVom(t *testing.T) {
	for _, test := range testCases {
		testName := test.Name + " mojo->vom"

		data, err := mojoEncode(test.MojoValue)
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

		out := reflect.New(reflect.TypeOf(test.MojoValue).Elem()).Interface()
		if err := mojoDecode(data, out); err != nil {
			t.Errorf("%s: error decoding mojo bytes %x: %v", testName, data, err)
			continue
		}

		if got, want := out, test.MojoValue; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: result doesn't match expectation. got %#v, but want %#v", testName, got, want)
		}
	}
}

func mojoEncode(mojoValue interface{}) ([]byte, error) {
	payload, ok := mojoValue.(encodable)
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

func mojoDecode(b []byte, outValue interface{}) error {
	dec := bindings.NewDecoder(b, nil)
	payload, ok := outValue.(decodable)
	if !ok {
		return fmt.Errorf("type %T lacks an Decode() method", outValue)
	}
	return payload.Decode(dec)
}

type encodable interface {
	Encode(encoder *bindings.Encoder) error
}

type decodable interface {
	Decode(decoder *bindings.Decoder) error
}
