// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder_test

import (
	"mojo/public/interfaces/bindings/mojom_types"
	"mojom/tests/transcoder_testcases"
	"reflect"
	"testing"

	"v.io/v23/vdl"
	"v.io/x/mojo/transcoder"
)

func TestVdlAndMojoTypeConversion(t *testing.T) {
	// Create types.
	enumType := vdl.NamedType("v23proxy/tests/transcoder_testcases.TestEnum", vdl.EnumType("A", "B", "C"))
	basicStructType := vdl.NamedType("v23proxy/tests/transcoder_testcases.TestBasicStruct", vdl.StructType(vdl.Field{"Enum", enumType}, vdl.Field{"A", vdl.Int32Type}))

	builder := vdl.TypeBuilder{}
	strct := builder.Struct()
	strct.AppendField("Enum", enumType)
	namedStruct := builder.Named("v23proxy/tests/transcoder_testcases.TestCyclicStruct").AssignBase(strct)
	strct.AppendField("CyclicStruct", builder.Optional().AssignElem(namedStruct))
	strct.AppendField("A", vdl.Int32Type)
	builder.Build()
	cyclicStructType, err := namedStruct.Built()
	if err != nil {
		t.Fatalf("error building struct: %v", err)
	}

	tests := []struct {
		vdl   *vdl.Type
		mojom mojom_types.Type
		mp    map[string]mojom_types.UserDefinedType
	}{
		{
			vdl.BoolType,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Bool},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.ByteType,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Uint8},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.Uint16Type,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Uint16},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.Uint32Type,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Uint32},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.Uint64Type,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Uint64},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.Int8Type,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Int8},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.Int16Type,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Int16},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.Int32Type,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Int32},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.Int64Type,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Int64},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.Float32Type,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Float},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.Float64Type,
			&mojom_types.TypeSimpleType{mojom_types.SimpleType_Double},
			map[string]mojom_types.UserDefinedType{},
		},
		{
			vdl.StringType,
			&mojom_types.TypeStringType{mojom_types.StringType{false}},
			map[string]mojom_types.UserDefinedType{},
		},
		// ?string is currently disallowed in vdl, so skipping
		{
			vdl.ArrayType(3, vdl.Int64Type),
			&mojom_types.TypeArrayType{mojom_types.ArrayType{false, 3, &mojom_types.TypeSimpleType{mojom_types.SimpleType_Int64}}},
			map[string]mojom_types.UserDefinedType{},
		},
		// ?[3]int64 is currently disallowed in vdl, so skipping
		{
			vdl.ListType(vdl.Int64Type),
			&mojom_types.TypeArrayType{mojom_types.ArrayType{false, -1, &mojom_types.TypeSimpleType{mojom_types.SimpleType_Int64}}},
			map[string]mojom_types.UserDefinedType{},
		},
		// ?[]int64 is currently disallowed in vdl, so skipping
		{
			vdl.MapType(vdl.Int64Type, vdl.BoolType),
			&mojom_types.TypeMapType{mojom_types.MapType{false, &mojom_types.TypeSimpleType{mojom_types.SimpleType_Int64}, &mojom_types.TypeSimpleType{mojom_types.SimpleType_Bool}}},
			map[string]mojom_types.UserDefinedType{},
		},
		// ?map[int64]bool is currently disallowed in vdl, so skipping
		{
			enumType,
			&mojom_types.TypeTypeReference{mojom_types.TypeReference{Nullable: false, TypeKey: stringPtr("TYPE_KEY:v23proxy.tests.transcoder_testcases.TestEnum")}},
			map[string]mojom_types.UserDefinedType{
				"TYPE_KEY:v23proxy.tests.transcoder_testcases.TestEnum": transcoder_testcases.GetAllMojomTypeDefinitions()["TYPE_KEY:v23proxy.tests.transcoder_testcases.TestEnum"],
			},
		},
		{
			basicStructType,
			&mojom_types.TypeTypeReference{mojom_types.TypeReference{Nullable: false, TypeKey: stringPtr("TYPE_KEY:v23proxy.tests.transcoder_testcases.TestBasicStruct")}},
			map[string]mojom_types.UserDefinedType{
				"TYPE_KEY:v23proxy.tests.transcoder_testcases.TestBasicStruct": transcoder_testcases.GetAllMojomTypeDefinitions()["TYPE_KEY:v23proxy.tests.transcoder_testcases.TestBasicStruct"],
				"TYPE_KEY:v23proxy.tests.transcoder_testcases.TestEnum":        transcoder_testcases.GetAllMojomTypeDefinitions()["TYPE_KEY:v23proxy.tests.transcoder_testcases.TestEnum"],
			},
		},
		{
			cyclicStructType,
			&mojom_types.TypeTypeReference{mojom_types.TypeReference{Nullable: false, TypeKey: stringPtr("TYPE_KEY:v23proxy.tests.transcoder_testcases.TestCyclicStruct")}},
			map[string]mojom_types.UserDefinedType{
				"TYPE_KEY:v23proxy.tests.transcoder_testcases.TestCyclicStruct": transcoder_testcases.GetAllMojomTypeDefinitions()["TYPE_KEY:v23proxy.tests.transcoder_testcases.TestCyclicStruct"],
				"TYPE_KEY:v23proxy.tests.transcoder_testcases.TestEnum":         transcoder_testcases.GetAllMojomTypeDefinitions()["TYPE_KEY:v23proxy.tests.transcoder_testcases.TestEnum"],
			},
		},
	}

	for _, test := range tests {
		mojomtype, mp := transcoder.VDLToMojomType(test.vdl)

		// Note: Equality is only guaranteed if the casing matches up. Mojom no longer sends out UpperCamelCase values.
		if !reflect.DeepEqual(mojomtype, test.mojom) {
			t.Errorf("vdl type %v, when converted to mojo type was %#v. expected %#v", test.vdl, mojomtype, test.mojom)
		}
		// Note: Equality of structs is virtually impossible. The DeclarationData of the values in the map contains lots of information we cannot recover from vdl.
		//if !reflect.DeepEqual(mp, test.mp) {
		//	t.Errorf("vdl type %v, when converted to mojo type did not match expected user defined types. got %#v, expected %#v", test.vdl, mp, test.mp)
		//}
		for k, _ := range test.mp {
			if _, ok := mp[k]; !ok {
				t.Errorf("vdl type %v, when converted to mojo type did not create an entry for %s", test.vdl, k)
			}
		}
		if len(mp) != len(test.mp) {
			t.Errorf("vdl type %v, when converted to mojo type created %d map entries, expected %d", test.vdl, len(mp), len(test.mp))
		}

		vt, err := transcoder.MojomToVDLType(test.mojom, test.mp)
		if err != nil {
			t.Errorf("error converting mojo type %#v (with user defined types %v): %v", test.mojom, test.mp, err)
		}
		if !reflect.DeepEqual(vt, test.vdl) {
			t.Errorf("mojom type %#v (with user defined types %v), when converted to vdl type was %v. expected %v", test.mojom, test.mp, vt, test.vdl)
		}
	}
}
