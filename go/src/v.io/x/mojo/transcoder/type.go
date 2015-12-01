// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"fmt"

	"mojo/public/interfaces/bindings/mojom_types"

	"v.io/v23/vdl"
)

/*// Given a descriptor mapping, produce 2 maps.
// The former maps from mojom identifiers to VDL Type.
// The latter maps from VDL Type string (hash cons) to the mojom identifier.
// These maps are used to interconvert more easily.
func AnalyzeMojomDescriptors(mp map[string]mojom_types.UserDefinedType) map[string]*vdl.Type {
	m2V := make(map[string]*vdl.Type)
	for s, udt := range mp {
		m2V[s] = mojomToVDLTypeUDT(udt, mp)
	}
	return m2V
}*/

// Convert the known type reference to a vdl type.
// Panics if the type reference was not known.
/*func TypeReferenceToVDLType(tr mojom_types.TypeReference, mp map[string]mojom_types.UserDefinedType) *vdl.Type {
	if udt, ok := mp[tr.TypeKey]; ok {
		return mojomToVDLTypeUDT(udt, mp)
	}
	panic("Type Key %s was not present in the mapping", tr.typeKey)
}*/

func mojomToVDLTypeUDT(udt mojom_types.UserDefinedType, mp map[string]mojom_types.UserDefinedType) (vt *vdl.Type) {
	u := interface{}(udt)
	switch u := u.(type) { // To do the type switch, udt has to be converted to interface{}.
	case *mojom_types.UserDefinedTypeEnumType: // enum
		me := u.Value

		// TODO: Assumes that the maximum enum index is len(me.Values) - 1.
		labels := make([]string, len(me.Values))
		for _, ev := range me.Values { // per EnumValue...
			// EnumValue has DeclData, EnumTypeKey, and IntValue.
			// We just need the first and last.
			labels[int(ev.IntValue)] = *ev.DeclData.ShortName
		}

		vt = vdl.NamedType(*me.DeclData.ShortName, vdl.EnumType(labels...))
	case *mojom_types.UserDefinedTypeStructType: // struct
		ms := u.Value

		vt = MojomStructToVDLType(ms, mp)
	case *mojom_types.UserDefinedTypeUnionType: // union
		mu := u.Value

		vfields := make([]vdl.Field, len(mu.Fields))
		for ix, mfield := range mu.Fields {
			vfields[ix] = vdl.Field{
				Name: *mfield.DeclData.ShortName,
				Type: MojomToVDLType(mfield.Type, mp),
			}
		}
		vt = vdl.NamedType(*mu.DeclData.ShortName, vdl.UnionType(vfields...))
	case *mojom_types.UserDefinedTypeInterfaceType: // interface
		panic("interfaces don't exist in vdl")
	default: // unknown
		panic(fmt.Errorf("user defined type %#v with unknown tag %d", udt, udt.Tag()))
	}
	return vt
}

func MojomStructToVDLType(ms mojom_types.MojomStruct, mp map[string]mojom_types.UserDefinedType) (vt *vdl.Type) {
	vfields := make([]vdl.Field, len(ms.Fields))
	for ix, mfield := range ms.Fields {
		vfields[ix] = vdl.Field{
			Name: *mfield.DeclData.ShortName,
			Type: MojomToVDLType(mfield.Type, mp),
		}
	}
	vt = vdl.NamedType(*ms.DeclData.ShortName, vdl.StructType(vfields...))
	return vt
}

// Given a mojom Type and the descriptor mapping, produce the corresponding vdltype.
func MojomToVDLType(mojomtype mojom_types.Type, mp map[string]mojom_types.UserDefinedType) (vt *vdl.Type) {
	// TODO(alexfandrianto): Cyclic types?
	mt := interface{}(mojomtype)
	switch mt := interface{}(mt).(type) { // To do the type switch, mt has to be converted to interface{}.
	case *mojom_types.TypeSimpleType: // TypeSimpleType
		switch mt.Value {
		case mojom_types.SimpleType_Bool:
			vt = vdl.BoolType
		case mojom_types.SimpleType_Double:
			vt = vdl.Float64Type
		case mojom_types.SimpleType_Float:
			vt = vdl.Float32Type
		case mojom_types.SimpleType_InT8:
			vt = vdl.Int8Type
		case mojom_types.SimpleType_InT16:
			vt = vdl.Int16Type
		case mojom_types.SimpleType_InT32:
			vt = vdl.Int32Type
		case mojom_types.SimpleType_InT64:
			vt = vdl.Int64Type
		case mojom_types.SimpleType_UinT8:
			vt = vdl.ByteType
		case mojom_types.SimpleType_UinT16:
			vt = vdl.Uint16Type
		case mojom_types.SimpleType_UinT32:
			vt = vdl.Uint32Type
		case mojom_types.SimpleType_UinT64:
			vt = vdl.Uint64Type
		}
	case *mojom_types.TypeStringType: // TypeStringType
		st := mt.Value
		if st.Nullable {
			panic("nullable strings don't exist in vdl")
		}
		vt = vdl.StringType
	case *mojom_types.TypeArrayType: // TypeArrayType
		at := mt.Value
		if at.Nullable {
			panic("nullable arrays don't exist in vdl")
		}
		if at.FixedLength > 0 {
			vt = vdl.ArrayType(int(at.FixedLength), MojomToVDLType(at.ElementType, mp))
		} else {
			vt = vdl.ListType(MojomToVDLType(at.ElementType, mp))
		}
	case *mojom_types.TypeMapType: // TypeMapType
		// Note that mojom doesn't have sets.
		m := mt.Value
		if m.Nullable {
			panic("nullable maps don't exist in vdl")
		}
		vt = vdl.MapType(MojomToVDLType(m.KeyType, mp), MojomToVDLType(m.ValueType, mp))
	case *mojom_types.TypeHandleType: // TypeHandleType
		panic("handles don't exist in vdl")
	case *mojom_types.TypeTypeReference: // TypeTypeReference
		tr := mt.Value
		if tr.IsInterfaceRequest {
			panic("interface requests don't exist in vdl")
		}
		udt := mp[*tr.TypeKey]
		if udt.Tag() != 1 && tr.Nullable {
			panic("nullable non-struct type reference cannot be represented in vdl")
		}
		vt = mojomToVDLTypeUDT(udt, mp)
	default:
		panic(fmt.Errorf("%#v has unknown tag %d", mojomtype, mojomtype.Tag()))
	}

	return vt
}

func VDLToMojomType(t *vdl.Type) (mojomtype mojom_types.Type, mp map[string]mojom_types.UserDefinedType) {
	mp = map[string]mojom_types.UserDefinedType{}
	mojomtype = vdlToMojomTypeInternal(t, false, mp)
	return
}

func vdlToMojomTypeInternal(t *vdl.Type, nullable bool, mp map[string]mojom_types.UserDefinedType) (mojomtype mojom_types.Type) {
	switch t.Kind() {
	case vdl.Bool, vdl.Float64, vdl.Float32, vdl.Int8, vdl.Int16, vdl.Int32, vdl.Int64, vdl.Byte, vdl.Uint16, vdl.Uint32, vdl.Uint64:
		return &mojom_types.TypeSimpleType{
			simpleTypeCode(t.Kind()),
		}
	case vdl.String:
		return &mojom_types.TypeStringType{
			stringType(nullable),
		}
	case vdl.Array:
		elem := vdlToMojomTypeInternal(t.Elem(), false, mp)
		return &mojom_types.TypeArrayType{
			arrayType(elem, nullable, t.Len()),
		}
	case vdl.List:
		elem := vdlToMojomTypeInternal(t.Elem(), false, mp)
		return &mojom_types.TypeArrayType{
			listType(elem, nullable),
		}
	case vdl.Map:
		key := vdlToMojomTypeInternal(t.Key(), false, mp)
		elem := vdlToMojomTypeInternal(t.Elem(), false, mp)
		return &mojom_types.TypeMapType{
			mapType(key, elem, nullable),
		}
	case vdl.Struct, vdl.Union, vdl.Enum:
		udtKey := userDefinedTypeKey(t, mp)
		return &mojom_types.TypeTypeReference{
			mojom_types.TypeReference{
				Nullable: nullable,
				TypeKey:  &udtKey,
			},
		}
	case vdl.Optional:
		return vdlToMojomTypeInternal(t.Elem(), true, mp)
	default:
		panic(fmt.Sprintf("conversion from VDL kind %v to mojom type not implemented", t.Kind()))
	}
}

func userDefinedTypeKey(t *vdl.Type, mp map[string]mojom_types.UserDefinedType) string {
	key := t.String()
	if _, ok := mp[key]; ok {
		return key
	}
	mp[key] = nil // placeholder to stop recursion

	var udt mojom_types.UserDefinedType
	switch t.Kind() {
	case vdl.Struct:
		udt = structType(t, mp)
	case vdl.Union:
		udt = unionType(t, mp)
	case vdl.Enum:
		udt = enumType(t)
	default:
		panic(fmt.Sprintf("conversion from VDL kind %v to mojom user defined type not implemented", t.Kind()))
	}

	mp[key] = udt
	return key
}

func simpleTypeCode(k vdl.Kind) mojom_types.SimpleType {
	switch k {
	case vdl.Bool:
		return mojom_types.SimpleType_Bool
	case vdl.Float64:
		return mojom_types.SimpleType_Double
	case vdl.Float32:
		return mojom_types.SimpleType_Float
	case vdl.Int8:
		return mojom_types.SimpleType_InT8
	case vdl.Int16:
		return mojom_types.SimpleType_InT16
	case vdl.Int32:
		return mojom_types.SimpleType_InT32
	case vdl.Int64:
		return mojom_types.SimpleType_InT64
	case vdl.Byte:
		return mojom_types.SimpleType_UinT8
	case vdl.Uint16:
		return mojom_types.SimpleType_UinT16
	case vdl.Uint32:
		return mojom_types.SimpleType_UinT32
	case vdl.Uint64:
		return mojom_types.SimpleType_UinT64
	default:
		panic(fmt.Sprintf("kind %v does not represent a simple type", k))
	}
}

func stringType(nullable bool) mojom_types.StringType {
	return mojom_types.StringType{nullable}
}

func arrayType(elem mojom_types.Type, nullable bool, length int) mojom_types.ArrayType {
	return mojom_types.ArrayType{nullable, int32(length), elem}
}

func listType(elem mojom_types.Type, nullable bool) mojom_types.ArrayType {
	return mojom_types.ArrayType{nullable, -1, elem}
}

func mapType(key, value mojom_types.Type, nullable bool) mojom_types.MapType {
	return mojom_types.MapType{nullable, key, value}
}

func structType(t *vdl.Type, mp map[string]mojom_types.UserDefinedType) mojom_types.UserDefinedType {
	layout := computeStructLayout(t)
	structFields := make([]mojom_types.StructField, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		byteOffset, _ := layout.MojoOffsetsFromVdlIndex(i)
		structFields[i] = mojom_types.StructField{
			Type:   vdlToMojomTypeInternal(t.Field(i).Type, false, mp),
			Offset: int32(byteOffset),
		}
	}
	return &mojom_types.UserDefinedTypeStructType{
		mojom_types.MojomStruct{
			Fields: structFields,
		},
	}
}

func unionType(t *vdl.Type, mp map[string]mojom_types.UserDefinedType) mojom_types.UserDefinedType {
	unionFields := make([]mojom_types.UnionField, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		unionFields[i] = mojom_types.UnionField{
			Type: vdlToMojomTypeInternal(t.Field(i).Type, false, mp),
			Tag:  uint32(i),
		}
	}
	return &mojom_types.UserDefinedTypeUnionType{
		mojom_types.MojomUnion{
			Fields: unionFields,
		},
	}
}

func enumType(t *vdl.Type) mojom_types.UserDefinedType {
	enumValues := make([]mojom_types.EnumValue, t.NumEnumLabel())
	for i := 0; i < t.NumEnumLabel(); i++ {
		enumValues[i] = mojom_types.EnumValue{
			EnumTypeKey: t.EnumLabel(i),
			IntValue:    int32(i),
		}
	}
	return &mojom_types.UserDefinedTypeEnumType{
		mojom_types.MojomEnum{
			Values: enumValues,
		},
	}
}
