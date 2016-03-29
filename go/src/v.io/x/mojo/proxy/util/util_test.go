// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util_test

import (
	"reflect"
	"testing"

	"v.io/v23/vdl"
	"v.io/v23/vom"
	"v.io/x/mojo/proxy/util"
)

type TestStructA struct {
	A int8
	B uint64
	C string
}

func TestSplitRawBytesStruct(t *testing.T) {
	inputStruct := TestStructA{
		1,
		2,
		"3",
	}
	expectedFieldsRaw := []*vom.RawBytes{
		vom.RawBytesOf(int8(1)),
		vom.RawBytesOf(uint64(2)),
		vom.RawBytesOf("3"),
	}

	target := util.StructSplitTarget()
	if err := vdl.FromReflect(target, reflect.ValueOf(inputStruct)); err != nil {
		t.Fatalf("error splitting target: %v", err)
	}

	if got, want := target.Fields(), expectedFieldsRaw; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestJoinRawBytesAsStruct(t *testing.T) {
	expectedStruct := TestStructA{
		1,
		2,
		"3",
	}
	expectedRaw := vom.RawBytesOf(expectedStruct)
	fieldsRaw := []*vom.RawBytes{
		vom.RawBytesOf(int8(1)),
		vom.RawBytesOf(uint64(2)),
		vom.RawBytesOf("3"),
	}

	var out *vom.RawBytes
	target, err := vdl.ReflectTarget(reflect.ValueOf(&out))
	if err != nil {
		t.Fatalf("error in ReflectTarget: %v", err)
	}
	if err := util.JoinRawBytesAsStruct(target, expectedRaw.Type, fieldsRaw); err != nil {
		t.Fatalf("error joining raw bytes: %v", err)
	}

	if got, want := out, expectedRaw; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
