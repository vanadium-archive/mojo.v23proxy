// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"mojo/public/interfaces/bindings/mojom_types"

	"v.io/v23/vdl"
)

// upperCamelCase converts thisString to ThisString.
func upperCamelCase(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

func MojomStructToVDLType(ms mojom_types.MojomStruct, mp map[string]mojom_types.UserDefinedType) (*vdl.Type, error) {
	builder := &vdl.TypeBuilder{}
	// Note: The type key is "" below because if there is a cycle, it will have a separate reference under a separate
	// type key and if there isn't the key is irrelevant.
	pending := mojomStructToVDLType("", ms, mp, builder, map[string]vdl.TypeOrPending{})
	builder.Build()
	return pending.Built()
}

func MojomToVDLType(mt mojom_types.Type, mp map[string]mojom_types.UserDefinedType) (*vdl.Type, error) {
	builder := &vdl.TypeBuilder{}
	t := mojomToVDLType(mt, mp, builder, map[string]vdl.TypeOrPending{})
	builder.Build()
	if vt, ok := t.(*vdl.Type); ok {
		return vt, nil
	}
	return t.(vdl.PendingType).Built()

}

func mojomStructToVDLType(typeKey string, ms mojom_types.MojomStruct, mp map[string]mojom_types.UserDefinedType, builder *vdl.TypeBuilder, pendingUdts map[string]vdl.TypeOrPending) (vt vdl.PendingType) {
	strct := builder.Struct()
	if ms.DeclData.FullIdentifier != nil {
		vt = builder.Named(mojomToVdlPath(*ms.DeclData.FullIdentifier)).AssignBase(strct)
	} else {
		vt = strct
	}
	pendingUdts[typeKey] = vt
	for _, mfield := range ms.Fields {
		strct.AppendField(upperCamelCase(*mfield.DeclData.ShortName), mojomToVDLType(mfield.Type, mp, builder, pendingUdts))
	}
	return
}

func mojomToVDLTypeUDT(typeKey string, udt mojom_types.UserDefinedType, mp map[string]mojom_types.UserDefinedType, builder *vdl.TypeBuilder, pendingUdts map[string]vdl.TypeOrPending) (vt vdl.TypeOrPending) {
	u := interface{}(udt)
	switch u := u.(type) { // To do the type switch, udt has to be converted to interface{}.
	case *mojom_types.UserDefinedTypeEnumType: // enum
		me := u.Value

		// TODO: Assumes that the maximum enum index is len(me.Values) - 1.
		labels := make([]string, len(me.Values))
		for _, ev := range me.Values { // per EnumValue...
			// EnumValue has DeclData, EnumTypeKey, and IntValue.
			// We just need the first and last.
			labels[int(ev.IntValue)] = upperCamelCase(*ev.DeclData.ShortName)
		}

		vt = vdl.NamedType(mojomToVdlPath(*me.DeclData.FullIdentifier), vdl.EnumType(labels...))
		pendingUdts[typeKey] = vt
	case *mojom_types.UserDefinedTypeStructType: // struct
		vt = mojomStructToVDLType(typeKey, u.Value, mp, builder, pendingUdts)
	case *mojom_types.UserDefinedTypeUnionType: // union
		mu := u.Value

		union := builder.Union()
		vt = builder.Named(mojomToVdlPath(*mu.DeclData.FullIdentifier)).AssignBase(union)
		pendingUdts[typeKey] = vt
		for _, mfield := range mu.Fields {
			union = union.AppendField(upperCamelCase(*mfield.DeclData.ShortName), mojomToVDLType(mfield.Type, mp, builder, pendingUdts))
		}
	case *mojom_types.UserDefinedTypeInterfaceType: // interface
		panic("interfaces don't exist in vdl")
	default: // unknown
		panic(fmt.Errorf("user defined type %#v with unknown tag %d", udt, udt.Tag()))
	}
	return vt
}

// Given a mojom Type and the descriptor mapping, produce the corresponding vdltype.
func mojomToVDLType(mojomtype mojom_types.Type, mp map[string]mojom_types.UserDefinedType, builder *vdl.TypeBuilder, pendingUdts map[string]vdl.TypeOrPending) (vt vdl.TypeOrPending) {
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
		case mojom_types.SimpleType_Int8:
			vt = vdl.Int8Type
		case mojom_types.SimpleType_Int16:
			vt = vdl.Int16Type
		case mojom_types.SimpleType_Int32:
			vt = vdl.Int32Type
		case mojom_types.SimpleType_Int64:
			vt = vdl.Int64Type
		case mojom_types.SimpleType_Uint8:
			vt = vdl.ByteType
		case mojom_types.SimpleType_Uint16:
			vt = vdl.Uint16Type
		case mojom_types.SimpleType_Uint32:
			vt = vdl.Uint32Type
		case mojom_types.SimpleType_Uint64:
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
			vt = builder.Array().
				AssignLen(int(at.FixedLength)).
				AssignElem(mojomToVDLType(at.ElementType, mp, builder, pendingUdts))
		} else {
			vt = builder.List().
				AssignElem(mojomToVDLType(at.ElementType, mp, builder, pendingUdts))
		}
	case *mojom_types.TypeMapType: // TypeMapType
		// Note that mojom doesn't have sets.
		m := mt.Value
		if m.Nullable {
			panic("nullable maps don't exist in vdl")
		}
		vt = builder.Map().
			AssignKey(mojomToVDLType(m.KeyType, mp, builder, pendingUdts)).
			AssignElem(mojomToVDLType(m.ValueType, mp, builder, pendingUdts))
	case *mojom_types.TypeHandleType: // TypeHandleType
		panic("handles don't exist in vdl")
	case *mojom_types.TypeTypeReference: // TypeTypeReference
		tr := mt.Value
		if tr.IsInterfaceRequest {
			panic("interface requests don't exist in vdl")
		}
		udt := mp[*tr.TypeKey]
		var ok bool
		vt, ok = pendingUdts[*tr.TypeKey]
		if !ok {
			vt = mojomToVDLTypeUDT(*tr.TypeKey, udt, mp, builder, pendingUdts)
		}
		if tr.Nullable {
			if udt.Tag() != 1 {
				panic("nullable non-struct type reference cannot be represented in vdl")
			}
			vt = builder.Optional().AssignElem(vt)
		}
	default:
		panic(fmt.Errorf("%#v has unknown tag %d", mojomtype, mojomtype.Tag()))
	}

	return vt
}

func VDLToMojomType(t *vdl.Type) (mojomtype mojom_types.Type, mp map[string]mojom_types.UserDefinedType) {
	mp = map[string]mojom_types.UserDefinedType{}
	mojomtype = vdlToMojomTypeInternal(t, true, false, mp)
	return
}

func vdlToMojomTypeInternal(t *vdl.Type, outermostType bool, nullable bool, mp map[string]mojom_types.UserDefinedType) (mojomtype mojom_types.Type) {
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
		elem := vdlToMojomTypeInternal(t.Elem(), false, false, mp)
		return &mojom_types.TypeArrayType{
			arrayType(elem, nullable, t.Len()),
		}
	case vdl.List:
		elem := vdlToMojomTypeInternal(t.Elem(), false, false, mp)
		return &mojom_types.TypeArrayType{
			listType(elem, nullable),
		}
	case vdl.Map:
		key := vdlToMojomTypeInternal(t.Key(), false, false, mp)
		elem := vdlToMojomTypeInternal(t.Elem(), false, false, mp)
		return &mojom_types.TypeMapType{
			mapType(key, elem, nullable),
		}
	case vdl.Struct, vdl.Union, vdl.Enum:
		udtKey := addUserDefinedType(t, mp)
		ret := &mojom_types.TypeTypeReference{
			mojom_types.TypeReference{
				Nullable: nullable,
				TypeKey:  &udtKey,
			},
		}
		if !outermostType {
			// This is needed to match the output of the generator exactly, the outermost type
			// is not given an identifier.
			ret.Value.Identifier = ret.Value.TypeKey
		}
		return ret
	case vdl.Optional:
		return vdlToMojomTypeInternal(t.Elem(), false, true, mp)
	default:
		panic(fmt.Sprintf("conversion from VDL kind %v to mojom type not implemented", t.Kind()))
	}
}

func addUserDefinedType(t *vdl.Type, mp map[string]mojom_types.UserDefinedType) string {
	key := mojomTypeKey(t)
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
		return mojom_types.SimpleType_Int8
	case vdl.Int16:
		return mojom_types.SimpleType_Int16
	case vdl.Int32:
		return mojom_types.SimpleType_Int32
	case vdl.Int64:
		return mojom_types.SimpleType_Int64
	case vdl.Byte:
		return mojom_types.SimpleType_Uint8
	case vdl.Uint16:
		return mojom_types.SimpleType_Uint16
	case vdl.Uint32:
		return mojom_types.SimpleType_Uint32
	case vdl.Uint64:
		return mojom_types.SimpleType_Uint64
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
	structFields := make([]mojom_types.StructField, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		structFields[i] = mojom_types.StructField{
			DeclData: &mojom_types.DeclarationData{ShortName: strPtr(t.Field(i).Name)},
			Type:     vdlToMojomTypeInternal(t.Field(i).Type, false, false, mp),
			Offset:   0, // Despite the fact that we can calculated the offset, set it to zero to match the generator
		}
	}
	_, name := vdl.SplitIdent(t.Name())
	return &mojom_types.UserDefinedTypeStructType{
		mojom_types.MojomStruct{
			DeclData: &mojom_types.DeclarationData{
				ShortName:      strPtr(name),
				FullIdentifier: strPtr(mojomIdentifier(t)),
			},
			Fields: structFields,
		},
	}
}

func unionType(t *vdl.Type, mp map[string]mojom_types.UserDefinedType) mojom_types.UserDefinedType {
	unionFields := make([]mojom_types.UnionField, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		unionFields[i] = mojom_types.UnionField{
			DeclData: &mojom_types.DeclarationData{ShortName: strPtr(t.Field(i).Name)},
			Type:     vdlToMojomTypeInternal(t.Field(i).Type, false, false, mp),
			Tag:      uint32(i),
		}
	}
	_, name := vdl.SplitIdent(t.Name())
	return &mojom_types.UserDefinedTypeUnionType{
		mojom_types.MojomUnion{
			DeclData: &mojom_types.DeclarationData{
				ShortName:      strPtr(name),
				FullIdentifier: strPtr(mojomIdentifier(t)),
			},
			Fields: unionFields,
		},
	}
}

func enumType(t *vdl.Type) mojom_types.UserDefinedType {
	enumValues := make([]mojom_types.EnumValue, t.NumEnumLabel())
	for i := 0; i < t.NumEnumLabel(); i++ {
		enumValues[i] = mojom_types.EnumValue{
			DeclData: &mojom_types.DeclarationData{ShortName: strPtr(t.EnumLabel(i))},
			IntValue: int32(i),
		}
	}
	_, name := vdl.SplitIdent(t.Name())
	return &mojom_types.UserDefinedTypeEnumType{
		mojom_types.MojomEnum{
			DeclData: &mojom_types.DeclarationData{
				ShortName:      strPtr(name),
				FullIdentifier: strPtr(mojomIdentifier(t)),
			},
			Values: enumValues,
		},
	}
}

func strPtr(x string) *string {
	return &x
}

// mojomTypeKey creates a key from the vdl type's name that matches the generator's key.
// The reason for exactly matching the generator is to simplify the tests.
func mojomTypeKey(t *vdl.Type) string {
	return fmt.Sprintf("TYPE_KEY:%s", mojomIdentifier(t))
}

func mojomIdentifier(t *vdl.Type) string {
	return strings.Replace(t.Name(), "/", ".", -1)
}

// "a.b.c.D" -> "a/b/c.D"
func mojomToVdlPath(path string) string {
	lastDot := strings.LastIndex(path, ".")
	if lastDot == -1 {
		return path
	}
	return strings.Replace(path[:lastDot], ".", "/", -1) + path[lastDot:]
}
