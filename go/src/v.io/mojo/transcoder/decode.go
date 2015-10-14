package mojovdl

import (
	"fmt"
	"reflect"

	"mojo/public/go/bindings"

	"v.io/v23/vdl"
)

// Decode decodes the mojom-encoded data into valptr, which must be a pointer to
// the desired value.  The datatype describes the type of the encoded data.
// Returns an error if the data cannot be decoded into valptr, based on the VDL
// value conversion rules.
func Decode(data []byte, datatype *vdl.Type, valptr interface{}) error {
	target, err := vdl.ReflectTarget(reflect.ValueOf(valptr))
	if err != nil {
		return err
	}
	d := &decoder{dec: bindings.NewDecoder(data, nil)}
	return d.decodeValue(datatype, target, true, false)
}

// DecodeValue is like Decode, but decodes mojom-encoded data into a vdl.Value.
func DecodeValue(data []byte, datatype *vdl.Type) (*vdl.Value, error) {
	v := new(vdl.Value)
	if err := Decode(data, datatype, &v); err != nil {
		return nil, err
	}
	return v, nil
}

type decoder struct {
	dec       *bindings.Decoder
	typeStack []*vdl.Type
}

func (d *decoder) decodeValue(vt *vdl.Type, target vdl.Target, isTopType, isNullable bool) error {
	switch vt.Kind() {
	case vdl.Bool:
		value, err := d.dec.ReadBool()
		if err != nil {
			return err
		}
		return target.FromBool(value, vt)
	case vdl.Int16:
		value, err := d.dec.ReadInt16()
		if err != nil {
			return err
		}
		return target.FromInt(int64(value), vt)
	case vdl.Int32:
		value, err := d.dec.ReadInt32()
		if err != nil {
			return err
		}
		return target.FromInt(int64(value), vt)
	case vdl.Int64:
		value, err := d.dec.ReadInt64()
		if err != nil {
			return err
		}
		return target.FromInt(value, vt)
	case vdl.Byte:
		value, err := d.dec.ReadUint8()
		if err != nil {
			return err
		}
		return target.FromUint(uint64(value), vt)
	case vdl.Uint16:
		value, err := d.dec.ReadUint16()
		if err != nil {
			return err
		}
		return target.FromUint(uint64(value), vt)
	case vdl.Uint32:
		value, err := d.dec.ReadUint32()
		if err != nil {
			return err
		}
		return target.FromUint(uint64(value), vt)
	case vdl.Uint64:
		value, err := d.dec.ReadUint64()
		if err != nil {
			return err
		}
		return target.FromUint(value, vt)
	case vdl.Float32:
		value, err := d.dec.ReadFloat32()
		if err != nil {
			return err
		}
		return target.FromFloat(float64(value), vt)
	case vdl.Float64:
		value, err := d.dec.ReadFloat64()
		if err != nil {
			return err
		}
		return target.FromFloat(value, vt)
	case vdl.String:
		switch ptr, err := d.dec.ReadPointer(); {
		case err != nil:
			return err
		case ptr == 0:
			return fmt.Errorf("invalid null string pointer")
		default:
			value, err := d.dec.ReadString()
			if err != nil {
				return err
			}
			return target.FromString(value, vt)
		}
		return nil
	case vdl.Enum:
		index, err := d.dec.ReadInt32()
		if err != nil {
			return err
		}
		if int(index) >= vt.NumEnumLabel() || index < 0 {
			return fmt.Errorf("enum label out of range")
		}
		target.FromEnumLabel(vt.EnumLabel(int(index)), vt)
		return nil
	case vdl.Complex64:
		panic("unimplemented")
	case vdl.Complex128:
		panic("unimplemented")
	case vdl.Array, vdl.List:
		switch ptr, err := d.dec.ReadPointer(); {
		case err != nil:
			return err
		case ptr == 0 && isNullable:
			return target.FromNil(vdl.OptionalType(vt))
		case ptr == 0 && !isNullable:
			return fmt.Errorf("invalid null struct pointer")
		}

		if vt.IsBytes() {
			str, err := d.dec.ReadString()
			if err != nil {
				return err
			}
			return target.FromBytes([]byte(str), vt)
		} else {
			elemBitSize := baseTypeSizeBits(vt.Elem())
			numElems, err := d.dec.StartArray(elemBitSize)
			if err != nil {
				return err
			}
			listTarget, err := target.StartList(vt, int(numElems))
			if err != nil {
				return err
			}
			for i := 0; i < int(numElems); i++ {
				elemTarget, err := listTarget.StartElem(i)
				if err != nil {
					return err
				}
				if err := d.decodeValue(vt.Elem(), elemTarget, false, false); err != nil {
					return err
				}
				if err := listTarget.FinishElem(elemTarget); err != nil {
					return err
				}
			}
			if err := target.FinishList(listTarget); err != nil {
				return err
			}
		}
		return d.dec.Finish()
	case vdl.Set:
		panic("unimplemented")
		/*switch ptr, err := d.dec.ReadPointer(); {
		case err != nil:
			return err
		case ptr == 0 && isNullable:
			return target.FromNil(vdl.OptionalType(vt))
		case ptr == 0 && !isNullable:
			return fmt.Errorf("invalid null struct pointer")
		}
		keyBitSize := baseTypeSizeBits(vt.Key())
		numKeys, err := d.dec.StartArray(keyBitSize)
		if err != nil {
			return err
		}
		setTarget, err := target.StartSet(vt, int(numKeys))
		if err != nil {
			return err
		}
		for i := 0; i < int(numKeys); i++ {
			keyTarget, err := setTarget.StartKey()
			if err != nil {
				return err
			}
			if err := d.decodeValue(mt.Key, keyTarget, false, false); err != nil {
				return err
			}
			if err := setTarget.FinishKey(keyTarget); err != nil {
				return err
			}
		}
		if err := target.FinishSet(setTarget); err != nil {
			return err
		}
		return d.dec.Finish()*/
	case vdl.Map:
		switch ptr, err := d.dec.ReadPointer(); {
		case err != nil:
			return err
		case ptr == 0 && isNullable:
			return target.FromNil(vdl.OptionalType(vt))
		case ptr == 0 && !isNullable:
			return fmt.Errorf("invalid null struct pointer")
		}
		if err := d.dec.StartMap(); err != nil {
			return err
		}
		var keys, values []*vdl.Value
		keysTarget, err := vdl.ReflectTarget(reflect.ValueOf(&keys))
		if err != nil {
			return err
		}
		keysListType := vdl.ListType(vt.Key())
		if err := d.decodeValue(keysListType, keysTarget, false, false); err != nil {
			return err
		}
		valuesTarget, err := vdl.ReflectTarget(reflect.ValueOf(&values))
		if err != nil {
			return err
		}
		valuesListType := vdl.ListType(vt.Elem())
		if err := d.decodeValue(valuesListType, valuesTarget, false, false); err != nil {
			return err
		}

		if len(keys) != len(values) {
			return fmt.Errorf("values don't match keys")
		}
		mapTarget, err := target.StartMap(vt, len(keys))
		if err != nil {
			return err
		}
		for i, key := range keys {
			value := values[i]

			keyTarget, err := mapTarget.StartKey()
			if err != nil {
				return err
			}
			if err := vdl.FromValue(keyTarget, key); err != nil {
				return err
			}
			fieldTarget, err := mapTarget.FinishKeyStartField(keyTarget)
			if err != nil {
				return err
			}
			if err := vdl.FromValue(fieldTarget, value); err != nil {
				return err
			}
			if err := mapTarget.FinishField(keyTarget, fieldTarget); err != nil {
				return err
			}
		}
		if err := target.FinishMap(mapTarget); err != nil {
			return err
		}

		return d.dec.Finish()
	case vdl.Struct:
		// TODO(toddw): See the comment in encoder.mojomStructSize; we rely on the
		// fields to be presented in the canonical mojom field ordering.
		if !isTopType {
			switch ptr, err := d.dec.ReadPointer(); {
			case err != nil:
				return err
			case ptr == 0 && isNullable:
				return target.FromNil(vdl.OptionalType(vt))
			case ptr == 0 && !isNullable:
				return fmt.Errorf("invalid null struct pointer")
			}
		}
		_, err := d.dec.StartStruct()
		if err != nil {
			return err
		}
		targetFields, err := target.StartFields(vt)
		if err != nil {
			return err
		}
		for i := 0; i < vt.NumField(); i++ {
			mfield := vt.Field(i)
			switch vkey, vfield, err := targetFields.StartField(mfield.Name); {
			// TODO(toddw): Handle err == vdl.ErrFieldNoExist case?
			case err != nil:
				return err
			default:
				if err := d.decodeValue(mfield.Type, vfield, false, false); err != nil {
					return err
				}
				if err := targetFields.FinishField(vkey, vfield); err != nil {
					return err
				}
			}
		}
		// TODO(toddw): Fill in fields that weren't decoded with their zero value.
		if err := target.FinishFields(targetFields); err != nil {
			return err
		}
		return d.dec.Finish()
	case vdl.Union:
		size, tag, err := d.dec.ReadUnionHeader()
		if err != nil {
			return err
		}
		if size == 0 {
			d.dec.SkipUnionValue()
			return target.FromNil(vdl.OptionalType(vt))
		}
		if int(tag) >= vt.NumField() {
			return fmt.Errorf("union tag out of bounds")
		}
		fld := vt.Field(int(tag))
		targetFields, err := target.StartFields(vt)
		if err != nil {
			return err
		}
		vKey, vField, err := targetFields.StartField(fld.Name)
		if err != nil {
			return err
		}
		if fld.Type.Kind() == vdl.Union {
			switch ptr, err := d.dec.ReadPointer(); {
			case err != nil:
				return err
			case ptr == 0 && isNullable:
				return target.FromNil(vdl.OptionalType(fld.Type))
			case ptr == 0 && !isNullable:
				return fmt.Errorf("invalid null union pointer")
			}
			if err := d.dec.StartNestedUnion(); err != nil {
				return err
			}
		}
		if err := d.decodeValue(fld.Type, vField, false, false); err != nil {
			return err
		}
		if fld.Type.Kind() == vdl.Union {
			if err := d.dec.Finish(); err != nil {
				return err
			}
		}
		if err := targetFields.FinishField(vKey, vField); err != nil {
			return err
		}
		if err := target.FinishFields(targetFields); err != nil {
			return err
		}
		d.dec.FinishReadingUnionValue()
		return nil
	case vdl.Optional:
		return d.decodeValue(vt.Elem(), target, false, true)
	case vdl.Any:
		panic("unimplemented")
	//case vdl.TypeObject:
	//	panic("unimplemented")
	default:
		panic(fmt.Errorf("decodeValue unhandled vdl type: %v", vt))
	}
}
