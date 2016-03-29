// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"fmt"
	"reflect"

	"mojo/public/go/bindings"

	"v.io/v23/vdl"
)

// FromMojo decodes the mojom-encoded data into valptr, which must be a pointer to
// the desired value.  The datatype describes the type of the encoded data.
// Returns an error if the data cannot be decoded into valptr, based on the VDL
// value conversion rules.
// TODO(bprosnitz) Consider reimplementing this using mojom_type instead of vdl type
// so that we can take advantage of the struct offset in the mojom type.
func ValueFromMojo(valptr interface{}, data []byte, datatype *vdl.Type) error {
	target, err := vdl.ReflectTarget(reflect.ValueOf(valptr))
	if err != nil {
		return err
	}
	return FromMojo(target, data, datatype)
}

func FromMojo(target vdl.Target, data []byte, datatype *vdl.Type) error {
	mtv := &mojomToTargetTranscoder{modec: bindings.NewDecoder(data, nil)}
	return mtv.transcodeValue(datatype, target, true, false)
}

type mojomToTargetTranscoder struct {
	modec     *bindings.Decoder
	typeStack []*vdl.Type
}

func (mtv *mojomToTargetTranscoder) transcodeValue(vt *vdl.Type, target vdl.Target, isTopType, isNullable bool) error {
	switch vt.Kind() {
	case vdl.Bool:
		value, err := mtv.modec.ReadBool()
		if err != nil {
			return err
		}
		return target.FromBool(value, vt)
	case vdl.Int8:
		value, err := mtv.modec.ReadInt8()
		if err != nil {
			return err
		}
		return target.FromInt(int64(value), vt)
	case vdl.Int16:
		value, err := mtv.modec.ReadInt16()
		if err != nil {
			return err
		}
		return target.FromInt(int64(value), vt)
	case vdl.Int32:
		value, err := mtv.modec.ReadInt32()
		if err != nil {
			return err
		}
		return target.FromInt(int64(value), vt)
	case vdl.Int64:
		value, err := mtv.modec.ReadInt64()
		if err != nil {
			return err
		}
		return target.FromInt(value, vt)
	case vdl.Byte:
		value, err := mtv.modec.ReadUint8()
		if err != nil {
			return err
		}
		return target.FromUint(uint64(value), vt)
	case vdl.Uint16:
		value, err := mtv.modec.ReadUint16()
		if err != nil {
			return err
		}
		return target.FromUint(uint64(value), vt)
	case vdl.Uint32:
		value, err := mtv.modec.ReadUint32()
		if err != nil {
			return err
		}
		return target.FromUint(uint64(value), vt)
	case vdl.Uint64:
		value, err := mtv.modec.ReadUint64()
		if err != nil {
			return err
		}
		return target.FromUint(value, vt)
	case vdl.Float32:
		value, err := mtv.modec.ReadFloat32()
		if err != nil {
			return err
		}
		return target.FromFloat(float64(value), vt)
	case vdl.Float64:
		value, err := mtv.modec.ReadFloat64()
		if err != nil {
			return err
		}
		return target.FromFloat(value, vt)
	case vdl.String:
		switch ptr, err := mtv.modec.ReadPointer(); {
		case err != nil:
			return err
		case ptr == 0:
			return fmt.Errorf("invalid null string pointer")
		default:
			value, err := mtv.modec.ReadString()
			if err != nil {
				return err
			}
			return target.FromString(value, vt)
		}
		return nil
	case vdl.Enum:
		index, err := mtv.modec.ReadInt32()
		if err != nil {
			return err
		}
		if int(index) >= vt.NumEnumLabel() || index < 0 {
			return fmt.Errorf("enum label out of range")
		}
		target.FromEnumLabel(vt.EnumLabel(int(index)), vt)
		return nil
	case vdl.Array, vdl.List:
		switch ptr, err := mtv.modec.ReadPointer(); {
		case err != nil:
			return err
		case ptr == 0 && isNullable:
			return target.FromZero(vdl.OptionalType(vt))
		case ptr == 0 && !isNullable:
			return fmt.Errorf("invalid null struct pointer")
		}

		if vt.IsBytes() {
			str, err := mtv.modec.ReadString()
			if err != nil {
				return err
			}
			return target.FromBytes([]byte(str), vt)
		} else {
			elemBitSize := baseTypeSizeBits(vt.Elem())
			numElems, err := mtv.modec.StartArray(elemBitSize)
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
				if err := mtv.transcodeValue(vt.Elem(), elemTarget, false, false); err != nil {
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
		return mtv.modec.Finish()
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
		switch ptr, err := mtv.modec.ReadPointer(); {
		case err != nil:
			return err
		case ptr == 0 && isNullable:
			return target.FromZero(vdl.OptionalType(vt))
		case ptr == 0 && !isNullable:
			return fmt.Errorf("invalid null struct pointer")
		}
		if err := mtv.modec.StartMap(); err != nil {
			return err
		}
		var keys, values []*vdl.Value
		keysTarget, err := vdl.ReflectTarget(reflect.ValueOf(&keys))
		if err != nil {
			return err
		}
		keysListType := vdl.ListType(vt.Key())
		if err := mtv.transcodeValue(keysListType, keysTarget, false, false); err != nil {
			return err
		}
		valuesTarget, err := vdl.ReflectTarget(reflect.ValueOf(&values))
		if err != nil {
			return err
		}
		valuesListType := vdl.ListType(vt.Elem())
		if err := mtv.transcodeValue(valuesListType, valuesTarget, false, false); err != nil {
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

		return mtv.modec.Finish()
	case vdl.Struct:
		// TODO(toddw): See the comment in encoder.mojomStructSize; we rely on the
		// fields to be presented in the canonical mojom field ordering.
		if !isTopType {
			switch ptr, err := mtv.modec.ReadPointer(); {
			case err != nil:
				return err
			case ptr == 0 && isNullable:
				return target.FromZero(vdl.OptionalType(vt))
			case ptr == 0 && !isNullable:
				return fmt.Errorf("invalid null struct pointer")
			}
		}
		_, err := mtv.modec.StartStruct()
		if err != nil {
			return err
		}
		targetFields, err := target.StartFields(vt)
		if err != nil {
			return err
		}
		for _, alloc := range computeStructLayout(vt) {
			mfield := vt.Field(alloc.vdlStructIndex)
			switch vkey, vfield, err := targetFields.StartField(mfield.Name); {
			// TODO(toddw): Handle err == vdl.ErrFieldNoExist case?
			case err != nil:
				return err
			default:
				if err := mtv.transcodeValue(mfield.Type, vfield, false, false); err != nil {
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
		return mtv.modec.Finish()
	case vdl.Union:
		size, tag, err := mtv.modec.ReadUnionHeader()
		if err != nil {
			return err
		}
		if size == 0 {
			mtv.modec.SkipUnionValue()
			return target.FromZero(vdl.OptionalType(vt))
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
			switch ptr, err := mtv.modec.ReadPointer(); {
			case err != nil:
				return err
			case ptr == 0 && isNullable:
				return target.FromZero(vdl.OptionalType(fld.Type))
			case ptr == 0 && !isNullable:
				return fmt.Errorf("invalid null union pointer")
			}
			if err := mtv.modec.StartNestedUnion(); err != nil {
				return err
			}
		}
		if err := mtv.transcodeValue(fld.Type, vField, false, false); err != nil {
			return err
		}
		if fld.Type.Kind() == vdl.Union {
			if err := mtv.modec.Finish(); err != nil {
				return err
			}
		}
		if err := targetFields.FinishField(vKey, vField); err != nil {
			return err
		}
		if err := target.FinishFields(targetFields); err != nil {
			return err
		}
		mtv.modec.FinishReadingUnionValue()
		return nil
	case vdl.Optional:
		return mtv.transcodeValue(vt.Elem(), target, false, true)
	case vdl.Any:
		panic("unimplemented")
	//case vdl.TypeObject:
	//	panic("unimplemented")
	default:
		panic(fmt.Errorf("decodeValue unhandled vdl type: %v", vt))
	}
}
