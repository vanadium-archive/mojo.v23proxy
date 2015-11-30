// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"testing"

	"reflect"

	"v.io/v23/vdl"
)

func TestComputeStructLayout(t *testing.T) {
	testCases := []struct {
		t      *vdl.Type
		layout structLayout
	}{
		{
			vdl.TypeOf(struct {
				A uint32
			}{}),
			structLayout{
				structLayoutField{0, 0, 0},
			},
		},
		{
			vdl.TypeOf(struct {
				A uint32
				B string
			}{}),
			structLayout{
				structLayoutField{0, 0, 0},
				structLayoutField{1, 8, 0},
			},
		},
		{
			vdl.TypeOf(struct {
				A uint32
				B string
				C uint32
			}{}),
			structLayout{
				structLayoutField{0, 0, 0},
				structLayoutField{2, 4, 0},
				structLayoutField{1, 8, 0},
			},
		},
		{
			vdl.TypeOf(struct {
				A uint32
				B string
				C bool
				D float32
			}{}),
			structLayout{
				structLayoutField{0, 0, 0},
				structLayoutField{2, 4, 0},
				structLayoutField{1, 8, 0},
				structLayoutField{3, 16, 0},
			},
		},
		{
			vdl.TypeOf(struct {
				A uint32
				C string
				B float32
				D bool
			}{}),
			structLayout{
				structLayoutField{0, 0, 0},
				structLayoutField{2, 4, 0},
				structLayoutField{1, 8, 0},
				structLayoutField{3, 16, 0},
			},
		},
		{
			vdl.TypeOf(struct {
				A uint16
				B bool
				C string
				D float32
				E bool
			}{}),
			structLayout{
				structLayoutField{0, 0, 0},
				structLayoutField{1, 2, 0},
				structLayoutField{4, 2, 1},
				structLayoutField{3, 4, 0},
				structLayoutField{2, 8, 0},
			},
		},
	}

	for _, test := range testCases {
		layout := computeStructLayout(test.t)
		if got, want := layout, test.layout; !reflect.DeepEqual(got, want) {
			t.Errorf("struct layout for type %v was %v but %v was expected", test.t, got, want)
		}
		for _, o := range test.layout {
			byteOffset, bitOffset := layout.MojoOffsetsFromVdlIndex(o.vdlStructIndex)
			if got, want := byteOffset, o.byteOffset; got != want {
				t.Errorf("byte offset doesn't match. got %v, want %v", got, want)
			}
			if got, want := bitOffset, o.bitOffset; got != want {
				t.Errorf("bit offset doesn't match. got %v, want %v", got, want)
			}
		}
	}
}
