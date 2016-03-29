// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"reflect"

	"v.io/v23/vdl"
)

// ToMojom encodes a value as mojom.
// This differs from the standard mojo encode because it uses the
// vdl package for reflection and can therefore handle RawBytes
// and vdl.Value.
func ToMojom(value interface{}) ([]byte, error) {
	vtm := ToMojomTarget()
	err := vdl.FromReflect(vtm, reflect.ValueOf(value))
	return vtm.Bytes(), err
}

// ToMojomTarget creates a vdl.Target that writes mojom bytes.
func ToMojomTarget() *targetToMojomTranscoder {
	return &targetToMojomTranscoder{
		allocator: &allocator{},
	}
}

type targetToMojomTranscoder struct {
	allocator *allocator
}

func (vtm *targetToMojomTranscoder) Bytes() []byte {
	return vtm.allocator.AllocatedBytes()
}

func (vtm *targetToMojomTranscoder) FromBool(src bool, tt *vdl.Type) error {
	panic("cannot encode top level bool")
}
func (vtm *targetToMojomTranscoder) FromUint(src uint64, tt *vdl.Type) error {
	panic("cannot encode top level uint")
}
func (vtm *targetToMojomTranscoder) FromInt(src int64, tt *vdl.Type) error {
	panic("cannot encode top level int")
}
func (vtm *targetToMojomTranscoder) FromFloat(src float64, tt *vdl.Type) error {
	panic("cannot encode top level float")
}
func (vtm *targetToMojomTranscoder) FromBytes(src []byte, tt *vdl.Type) error {
	panic("cannot encode top level bytes")
}
func (vtm *targetToMojomTranscoder) FromString(src string, tt *vdl.Type) error {
	panic("cannot encode top level string")
}
func (vtm *targetToMojomTranscoder) FromEnumLabel(src string, tt *vdl.Type) error {
	panic("cannot encode top level enum")
}
func (vtm *targetToMojomTranscoder) FromTypeObject(src *vdl.Type) error {
	panic("cannot encode top level type object")
}
func (vtm *targetToMojomTranscoder) FromZero(tt *vdl.Type) error {
	if tt.Kind() == vdl.Struct {
		st, err := vtm.StartFields(tt)
		if err != nil {
			return err
		}
		for i := 0; i < tt.NumField(); i++ {
			fld := tt.Field(i)
			kt, ft, err := st.StartField(fld.Name)
			if err != nil {
				return err
			}
			if err := ft.FromZero(fld.Type); err != nil {
				return err
			}
			if err := st.FinishField(kt, ft); err != nil {
				return err
			}
		}
		return vtm.FinishFields(st)
	}
	panic("UNIMPLEMENTED")
}

func (vtm *targetToMojomTranscoder) StartList(tt *vdl.Type, len int) (vdl.ListTarget, error) {
	panic("UNIMPLEMENTED")
	return nil, nil
}
func (vtm *targetToMojomTranscoder) FinishList(x vdl.ListTarget) error {
	return nil
}
func (vtm *targetToMojomTranscoder) StartSet(tt *vdl.Type, len int) (vdl.SetTarget, error) {
	panic("UNIMPLEMENTED")
}
func (vtm *targetToMojomTranscoder) FinishSet(x vdl.SetTarget) error {
	panic("UNIMPLEMENTED")

}
func (vtm *targetToMojomTranscoder) StartMap(tt *vdl.Type, len int) (vdl.MapTarget, error) {
	panic("UNIMPLEMENTED")

}
func (vtm *targetToMojomTranscoder) FinishMap(x vdl.MapTarget) error {
	panic("UNIMPLEMENTED")

}
func (vtm *targetToMojomTranscoder) StartFields(tt *vdl.Type) (vdl.FieldsTarget, error) {
	if tt.Kind() != vdl.Struct {
		// Top-level unions not currently supported
		panic("UNIMPLEMENTED")
	}
	fieldsTarget, _, err := structFieldShared(tt, vtm.allocator, false)
	return fieldsTarget, err
}

func (vtm *targetToMojomTranscoder) FinishFields(x vdl.FieldsTarget) error {
	return nil
}
