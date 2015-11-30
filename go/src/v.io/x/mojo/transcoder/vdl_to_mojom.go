// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"reflect"

	"v.io/v23/vdl"
)

func VdlToMojom(value interface{}) ([]byte, error) {
	vtm := &vdlToMojomTranscoder{
		allocator: &allocator{},
	}
	err := vdl.FromReflect(vtm, reflect.ValueOf(value))
	return vtm.Bytes(), err
}

type vdlToMojomTranscoder struct {
	allocator *allocator
}

func (vtm *vdlToMojomTranscoder) Bytes() []byte {
	return vtm.allocator.AllocatedBytes()
}

func (vtm *vdlToMojomTranscoder) FromBool(src bool, tt *vdl.Type) error {
	panic("cannot encode top level bool")
}
func (vtm *vdlToMojomTranscoder) FromUint(src uint64, tt *vdl.Type) error {
	panic("cannot encode top level uint")
}
func (vtm *vdlToMojomTranscoder) FromInt(src int64, tt *vdl.Type) error {
	panic("cannot encode top level int")
}
func (vtm *vdlToMojomTranscoder) FromFloat(src float64, tt *vdl.Type) error {
	panic("cannot encode top level float")
}
func (vtm *vdlToMojomTranscoder) FromComplex(src complex128, tt *vdl.Type) error {
	panic("cannot encode top level complex")
}
func (vtm *vdlToMojomTranscoder) FromBytes(src []byte, tt *vdl.Type) error {
	panic("cannot encode top level bytes")
}
func (vtm *vdlToMojomTranscoder) FromString(src string, tt *vdl.Type) error {
	panic("cannot encode top level string")
}
func (vtm *vdlToMojomTranscoder) FromEnumLabel(src string, tt *vdl.Type) error {
	panic("cannot encode top level enum")
}
func (vtm *vdlToMojomTranscoder) FromTypeObject(src *vdl.Type) error {
	panic("cannot encode top level type object")
}
func (vtm *vdlToMojomTranscoder) FromNil(tt *vdl.Type) error {
	panic("cannot encode top level nil")
}

func (vtm *vdlToMojomTranscoder) StartList(tt *vdl.Type, len int) (vdl.ListTarget, error) {
	panic("UNIMPLEMENTED")
	return nil, nil
}
func (vtm *vdlToMojomTranscoder) FinishList(x vdl.ListTarget) error {
	return nil
}
func (vtm *vdlToMojomTranscoder) StartSet(tt *vdl.Type, len int) (vdl.SetTarget, error) {
	panic("UNIMPLEMENTED")
}
func (vtm *vdlToMojomTranscoder) FinishSet(x vdl.SetTarget) error {
	panic("UNIMPLEMENTED")

}
func (vtm *vdlToMojomTranscoder) StartMap(tt *vdl.Type, len int) (vdl.MapTarget, error) {
	panic("UNIMPLEMENTED")

}
func (vtm *vdlToMojomTranscoder) FinishMap(x vdl.MapTarget) error {
	panic("UNIMPLEMENTED")

}
func (vtm *vdlToMojomTranscoder) StartFields(tt *vdl.Type) (vdl.FieldsTarget, error) {
	if tt.Kind() != vdl.Struct {
		// Top-level unions not currently supported
		panic("UNIMPLEMENTED")
	}
	fieldsTarget, _, err := structFieldShared(tt, vtm.allocator, false)
	return fieldsTarget, err
}

func (vtm *vdlToMojomTranscoder) FinishFields(x vdl.FieldsTarget) error {
	return nil
}
