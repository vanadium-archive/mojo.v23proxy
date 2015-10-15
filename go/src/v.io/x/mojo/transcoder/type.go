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
			// EnumValue contains a ConstantOccurrence
			// ConstantOccurrence contains a ConstantValue
			// ConstantValue is a union that contains an EnumConstantValue
			ecv := ev.Value.Value.(*mojom_types.ConstantValueEnumValue).Value

			// EnumConstantValue contains the EnumValueName and IntValue
			labels[int(ecv.IntValue)] = *ecv.EnumValueName
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
			Type: MojomToVDLType(mfield.FieldType, mp),
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
			vt = vdl.BoolType
		case mojom_types.SimpleType_Float:
			vt = vdl.BoolType
		case mojom_types.SimpleType_InT8:
			panic("int8 doesn't exist in vdl")
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

/*func V2M(vt *vdl.Type, v2M map[string]string) mojom_types.Type {
	if m, ok := v2M[vt.String()]; ok {
		return mojom_types.TypeTypeReference{
			Value: mojom_types.TypeReference{
				Nullable:   nullable,
				Identifier: m,
				TypeKey:    m,
			},
		}
	}
	panic("vdl type %#v was not present in the mapping", vt)
}

// From the vdltype and the reverse mapping of the descriptor (hashcons vdltype string => typekey),
// produce the corresponding mojom Type.
func VDLToMojomType(vt *vdl.Type, v2M map[string]string) mojom_types.Type {
	return vdlToMojomTypeImpl(vt, v2M, false)
}

func vdlToMojomTypeImpl(vt *vdl.Type, v2M map[string]string, bool nullable) mojom_types.Type {
	if m, ok := v2M[vt.String()]; ok {
		return mojom_types.TypeTypeReference{
			Value: mojom_types.TypeReference{
				Nullable:   nullable,
				Identifier: m,
				TypeKey:    m,
			},
		}
	}

	fmt.Println("Missed the vdl to mojom map")
	// In the unlikely case where v2M was insufficient, we have the remaining logic.

	switch vt.Kind() {
	case vdl.Bool:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_Bool}
	case vdl.Byte:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_UinT8}
	case vdl.Uint16:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_UinT16}
	case vdl.Uint32:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_UinT32}
	case vdl.Uint64:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_UinT64}
	case vdl.Int16:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_InT16}
	case vdl.Int32:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_InT32}
	case vdl.Int64:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_InT64}
	case vdl.Float32:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_Float}
	case vdl.Float64:
		return mojom_types.TypeSimpleType{Value: mojom_types.SimpleType_Double}
	case vdl.Complex64:
		panic("complex float doesn't exist in mojom")
	case vdl.Complex128:
		panic("complex double doesn't exist in mojom")
	case vdl.String:
		return mojom_types.TypeStringType{Value: mojom_types.StringType{}}
	case vdl.Array:
		elemType := VDLToMojomType(vt.Elem(), v2M)
		return mojom_types.TypeArrayType{
			Value: mojom_types.ArrayType{
				FixedLength: int64(vt.Len()),
				ElementType: elemType,
			},
		}
	case vdl.List:
		elemType := VDLToMojomType(vt.Elem(), v2M)
		return mojom_types.TypeArrayType{
			Value: mojom_types.ArrayType{
				FixedLength: -1,
				ElementType: elemType,
			},
		}
	case vdl.Set:
		panic("set doesn't exist in mojom")
	case vdl.Map:
		keyType := VDLToMojomType(vt.Key(), v2M)
		elemType := VDLToMojomType(vt.Elem(), v2M)
		return mojom_types.TypeMapType{
			Value: mojom_types.MapType{
				KeyType:   &keyType,
				ValueType: &elemType,
			},
		}
	case vdl.Struct, vdl.Union, vdl.Enum:
		mt := mojom_types.TypeTypeReference{
			Value: mojom_types.TypeReference{
				Nullable:   nullable,
				Identifier: v2M[vt.String()],
				TypeKey:    v2M[vt.String()],
			},
		}
		return mt
	case vdl.TypeObject:
		panic("typeobject doesn't exist in mojom")
	case vdl.Any:
		panic("any doesn't exist in mojom")
	case vdl.Optional:
		// TODO(alexfandrianto): Unfortunately, without changing vdl, we can only
		// manage optional (named) structs. This doesn't Nullify anything else.
		return vdlToMojomTypeImpl(vt.Elem(), v2M, true)
	}
	panic(fmt.Errorf("%v can't be converted to MojomType", vt))
}
*/
