// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder_test

import (
	"mojo/public/interfaces/bindings/tests/rect"
	"mojo/public/interfaces/bindings/tests/test_structs"
)

type transcodeTestCase struct {
	Name      string
	MojoValue interface{}
	VdlValue  interface{}
}

// Test cases for the mojom <-> vdl transcoder tests. See transcoder_test.go
var testCases = []transcodeTestCase{
	// from Mojo's rect
	{
		Name: "Rect",
		MojoValue: &rect.Rect{
			X:      11,
			Y:      12,
			Height: 13,
			Width:  14,
		},
		VdlValue: rect.Rect{
			X:      11,
			Y:      12,
			Height: 13,
			Width:  14,
		},
	},
	// from Mojo's test_structs
	{
		Name: "NamedRegion",
		MojoValue: &test_structs.NamedRegion{
			Name: stringPtr("A"),
			Rects: &[]rect.Rect{
				rect.Rect{},
			},
		},
		VdlValue: test_structs.NamedRegion{
			Name: stringPtr("A"),
			Rects: &[]rect.Rect{
				rect.Rect{},
			},
		},
	},
	{
		Name: "RectPair",
		MojoValue: &test_structs.RectPair{
			First:  &rect.Rect{X: 0, Y: 1, Height: 2, Width: 3},
			Second: &rect.Rect{X: 11, Y: 12, Height: 13, Width: 14},
		},
		VdlValue: test_structs.RectPair{
			First:  &rect.Rect{X: 0, Y: 1, Height: 2, Width: 3},
			Second: &rect.Rect{X: 11, Y: 12, Height: 13, Width: 14},
		},
	},
	{
		Name:      "EmptyStruct",
		MojoValue: &test_structs.EmptyStruct{},
		VdlValue:  test_structs.EmptyStruct{},
	},
	// TODO(bprosnitz) HandleStruct?
	// TODO(bprosnitz) NullableHandleStruct?
	// TODO(bprosnitz) NoDefaultFieldValues?
	// TODO(bprosnitz) DefaultFieldValues?
	{
		Name: "ScopedConstants",
		MojoValue: &test_structs.ScopedConstants{
			test_structs.ScopedConstants_EType_E0,
			test_structs.ScopedConstants_EType_E1,
			test_structs.ScopedConstants_EType_E2,
			test_structs.ScopedConstants_EType_E3,
			test_structs.ScopedConstants_EType_E4,
			10,
			10,
		},
		VdlValue: test_structs.ScopedConstants{
			test_structs.ScopedConstants_EType_E0,
			test_structs.ScopedConstants_EType_E1,
			test_structs.ScopedConstants_EType_E2,
			test_structs.ScopedConstants_EType_E3,
			test_structs.ScopedConstants_EType_E4,
			10,
			10,
		},
	},
	// TODO(bprosnitz) MapKeyTypes?
	// TODO(bprosnitz) MapValueTypes?
	// TODO(bprosnitz) ArrayValueTypes?
	/*
		{
			Name: "UnsignedArrayValueTypes",
			MojoValue: &test_structs.UnsignedArrayValueTypes{
				[]uint8{1}, []uint16{2}, []uint32{3}, []uint64{4}, []float32{5}, []float64{6},
			},
			VdlValue: test_structs.UnsignedArrayValueTypes{
				[]uint8{1}, []uint16{2}, []uint32{3}, []uint64{4}, []float32{5}, []float64{6},
			},
		},
		{
			Name: "UnsignedFixedArrayValueTypes",
			MojoValue: &test_structs.UnsignedFixedArrayValueTypes{
				[3]uint8{1}, [2]uint16{2}, [2]uint32{3}, [2]uint64{4}, [2]float32{5}, [2]float64{6},
			},
			VdlValue: test_structs.UnsignedFixedArrayValueTypes{
				[3]uint8{1}, [2]uint16{2}, [2]uint32{3}, [2]uint64{4}, [2]float32{5}, [2]float64{6},
			},
		},
		{
			Name: "BoolArrayValueTypes",
			MojoValue: &test_structs.BoolArrayValueTypes{
				[]bool{false, true, true, false},
			},
			VdlValue: test_structs.BoolArrayValueTypes{
				[]bool{false, true, true, false},
			},
		},*/
	{
		Name: "FloatNumberValues",
		MojoValue: &test_structs.FloatNumberValues{
			0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9,
		},
		VdlValue: test_structs.FloatNumberValues{
			0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9,
		},
	},
	// TODO(bprosnitz) IntegerNumberValues?
	{
		Name: "UnsignedNumberValues",
		MojoValue: &test_structs.UnsignedNumberValues{
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
		},
		VdlValue: test_structs.UnsignedNumberValues{
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
		},
	},
	{
		Name: "BitArrayValues",
		MojoValue: &test_structs.BitArrayValues{
			[1]bool{true}, [7]bool{true, false, true}, [9]bool{true, false, true}, []bool{true, false, true},
			[][]bool{[]bool{true, false, true}}, []*[]bool{&[]bool{true, false, true}}, []*[2]bool{&[2]bool{true, false}},
		},
		VdlValue: test_structs.BitArrayValues{
			[1]bool{true}, [7]bool{true, false, true}, [9]bool{true, false, true}, []bool{true, false, true},
			[][]bool{[]bool{true, false, true}}, []*[]bool{&[]bool{true, false, true}}, []*[2]bool{&[2]bool{true, false}},
		},
	},
	// TODO(bprosnitz) MultiVersionStruct? + other versions
	// from Mojo's test_unions
	// TODO(bprosnitz) PodUnion?
	// TODO(bprosnitz) ObjectUnion?
	// TODO(bprosnitz) HandleUnion?
	// TODO(bprosnitz) WrapperStruct?
	// TODO(bprosnitz) DummyStruct?
	// TODO(bprosnitz) SmallStruct?
	// TODO(bprosnitz) SmallStructNonNullableUnion?
	// TODO(bprosnitz) SmallObjStruct?
	// TODO(bprosnitz) TryNonNullStruct?
	// TODO(bprosnitz) OldUnion?
	// TODO(bprosnitz) NewUnion?
	// TODO(bprosnitz) IncludingStruct?
	// test cases not from Mojo:
	/*
		// TODO(bprosnitz) These fail:
		{
			Name:      "UnnamedPrimitiveTestStruct",
			MojoValue: &transcoder_testcases.UnnamedPrimitiveTestStruct{1, "A", true, 2},
			VdlValue:  transcoder_testcases.UnnamedPrimitiveTestStruct{1, "A", true, 2},
		},
		{
			Name:      "Transcode to Named Primitives",
			MojoValue: &transcoder_testcases.UnnamedPrimitiveTestStruct{1, "A", true, 2},
			VdlValue:  NamedPrimitiveTestStruct{1, "A", true, 2},
		},*/
	// TODO(bprosnitz) More tests of errors, named type conversions, unsupported types, etc
}

func stringPtr(in string) *string { return &in }

type NUint32 uint32
type NString string
type NBool bool
type NFloat32 float32
type NamedPrimitiveTestStruct struct {
	A NUint32
	B NString
	C NBool
	D NFloat32
}
