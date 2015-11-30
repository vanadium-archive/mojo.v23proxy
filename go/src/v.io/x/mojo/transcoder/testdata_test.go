// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder_test

import (
	"mojo/public/interfaces/bindings/tests/rect"
	"mojo/public/interfaces/bindings/tests/test_structs"
	"mojo/public/interfaces/bindings/tests/test_unions"

	"mojom/tests/transcoder_testcases"

	"v.io/x/mojo/transcoder/testtypes"
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
	{
		Name: "DefaultFieldValues",
		MojoValue: &test_structs.DefaultFieldValues{
			true, 100, 100, 100, 100, 100, 100, 100, 100,
			100, 100, 100, 100,
			"foo", stringPtr("foo"),
			rect.Rect{X: 0, Y: 1, Height: 2, Width: 3},
			&rect.Rect{X: 4, Y: 5, Height: 6, Width: 7},
		},
		VdlValue: test_structs.DefaultFieldValues{
			true, 100, 100, 100, 100, 100, 100, 100, 100,
			100, 100, 100, 100,
			"foo", stringPtr("foo"),
			rect.Rect{X: 0, Y: 1, Height: 2, Width: 3},
			&rect.Rect{X: 4, Y: 5, Height: 6, Width: 7},
		},
	},
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
	{
		Name: "MapKeyTypes",
		MojoValue: &test_structs.MapKeyTypes{
			map[bool]bool{true: false},
			map[int8]int8{-1: 1},
			map[uint8]uint8{1: 1},
			map[int16]int16{-2: 2},
			map[uint16]uint16{2: 2},
			map[int32]int32{-4: 4},
			map[uint32]uint32{4: 4},
			map[int64]int64{-8: 8},
			map[uint64]uint64{8: 8},
			map[float32]float32{0.1: 0.1},
			map[float64]float64{0.2: 0.2},
			map[string]string{"A": "B", "C": "D"},
		},
		VdlValue: test_structs.MapKeyTypes{
			map[bool]bool{true: false},
			map[int8]int8{-1: 1},
			map[uint8]uint8{1: 1},
			map[int16]int16{-2: 2},
			map[uint16]uint16{2: 2},
			map[int32]int32{-4: 4},
			map[uint32]uint32{4: 4},
			map[int64]int64{-8: 8},
			map[uint64]uint64{8: 8},
			map[float32]float32{0.1: 0.1},
			map[float64]float64{0.2: 0.2},
			map[string]string{"A": "B", "C": "D"},
		},
	},
	// TODO(bprosnitz) MapValueTypes?
	{
		Name: "ArrayValueTypes",
		MojoValue: &test_structs.ArrayValueTypes{
			[]int8{1},
			[]int16{1, 2},
			[]int32{1, 2, 3},
			[]int64{1, 2, 3, 4},
			[]float32{1},
			[]float64{1, 2},
		},
		VdlValue: test_structs.ArrayValueTypes{
			[]int8{1},
			[]int16{1, 2},
			[]int32{1, 2, 3},
			[]int64{1, 2, 3, 4},
			[]float32{1},
			[]float64{1, 2},
		},
	},
	{
		Name: "FloatNumberValues",
		MojoValue: &test_structs.FloatNumberValues{
			0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9,
		},
		VdlValue: test_structs.FloatNumberValues{
			0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9,
		},
	},
	{
		Name: "IntegerNumberValues",
		MojoValue: &test_structs.IntegerNumberValues{
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		},
		VdlValue: test_structs.IntegerNumberValues{
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		},
	},
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
	// TODO(bprosnitz) Multi-version structs are not yet supported because the version is specified in the
	// struct header in the mojom bytes and we don't have any way to specify the version with VDL.
	/*
		{
			Name: "MultiVersionStruct Full -> V3",
			MojoValue: &test_structs.MultiVersionStruct{
				FInt32:  8,
				FRect:   &rect.Rect{1, 2, 3, 4},
				FString: stringPtr("testStr"),
			},
			VdlValue: testtypes.MultiVersionStructV3{
				FInt32:  8,
				FRect:   testtypes.Rect{1, 2, 3, 4},
				FString: "testStr",
			},
		},
		{
				Name: "MultiVersionStruct V3 -> Full",
				MojoValue: &test_structs.MultiVersionStructV3{
					FInt32:  8,
					FRect:   &rect.Rect{1, 2, 3, 4},
					FString: stringPtr("testStr"),
				},
				VdlValue: testtypes.MultiVersionStruct{
					FInt32:  8,
					FRect:   testtypes.Rect{1, 2, 3, 4},
					FString: "testStr",
				},
			},*/
	// from Mojo's test_unions
	{
		Name:      "PodUnionFInt8",
		MojoValue: &transcoder_testcases.PodUnionWrapper{&test_unions.PodUnionFInt8{-1}},
		VdlValue:  testtypes.PodUnionWrapper{testtypes.PodUnionFInt8{-1}},
	},
	{
		Name:      "ObjectUnionFInt8",
		MojoValue: &transcoder_testcases.ObjectUnionWrapper{&test_unions.ObjectUnionFInt8{5}},
		VdlValue:  testtypes.ObjectUnionWrapper{testtypes.ObjectUnionFInt8{5}},
	},
	{
		Name:      "ObjectUnionFDummy",
		MojoValue: &transcoder_testcases.ObjectUnionWrapper{&test_unions.ObjectUnionFDummy{test_unions.DummyStruct{5}}},
		VdlValue:  testtypes.ObjectUnionWrapper{testtypes.ObjectUnionFDummy{testtypes.DummyStruct{5}}},
	},
	{
		Name:      "ObjectUnionFPodUnion",
		MojoValue: &transcoder_testcases.ObjectUnionWrapper{&test_unions.ObjectUnionFPodUnion{&test_unions.PodUnionFDouble{1}}},
		VdlValue:  testtypes.ObjectUnionWrapper{testtypes.ObjectUnionFPodUnion{testtypes.PodUnionFDouble{1}}},
	},
	// TODO(bprosnitz) HandleUnion?
	// TODO(bprosnitz) WrapperStruct?
	// TODO(bprosnitz) SmallStruct?
	// TODO(bprosnitz) SmallStructNonNullableUnion?
	// TODO(bprosnitz) SmallObjStruct?
	{
		Name:      "TryNonNullStruct - nil optional",
		MojoValue: &test_unions.TryNonNullStruct{},
		VdlValue:  test_unions.TryNonNullStruct{},
	},
	{
		Name:      "TryNonNullStruct - non nil optional",
		MojoValue: &test_unions.TryNonNullStruct{&test_unions.DummyStruct{1}, test_unions.DummyStruct{2}},
		VdlValue:  test_unions.TryNonNullStruct{&test_unions.DummyStruct{1}, test_unions.DummyStruct{2}},
	},
	// TODO(bprosnitz) OldUnion?
	// TODO(bprosnitz) NewUnion?
	// TODO(bprosnitz) IncludingStruct?
	// test cases not from Mojo:
	/* This doesn't currently work because VDL doesn't register the anonymous struct.
	{
		Name:      "Transcode to Anonymous Struct",
		MojoValue: &transcoder_testcases.UnnamedPrimitiveTestStruct{1, "A", true, 2},
		VdlValue: struct {
			A uint32
			B string
			C bool
			D float32
		}{1, "A", true, 2},
	},*/
	{
		Name:      "Transcode to struct of different name",
		MojoValue: &transcoder_testcases.UnnamedPrimitiveTestStruct{1, "A", true, 2},
		VdlValue:  NamedPrimitiveTestStruct{1, "A", true, 2},
	},
	{
		Name:      "VarietyOfBitSizesStruct",
		MojoValue: &transcoder_testcases.VarietyOfBitSizesStruct{false, 8, 16, 32, 64, "F", []int8{1, 2}, map[string]bool{"a": true}, 32, 16, 8, false, true, 12},
		VdlValue:  transcoder_testcases.VarietyOfBitSizesStruct{false, 8, 16, 32, 64, "F", []int8{1, 2}, map[string]bool{"a": true}, 32, 16, 8, false, true, 12},
	},
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
