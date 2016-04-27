// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"

	"v.io/v23/vdl"
	"v.io/v23/vom"
)

type structSplitTarget struct {
	tt     *vdl.Type
	fields []*vom.RawBytes
	vdl.Target
}

func (targ *structSplitTarget) StartFields(tt *vdl.Type) (vdl.FieldsTarget, error) {
	targ.tt = tt
	targ.fields = make([]*vom.RawBytes, tt.NumField())
	return &structSplitFieldsTarget{targ}, nil
}

func (targ *structSplitTarget) FinishFields(x vdl.FieldsTarget) error {
	return nil
}

func (targ *structSplitTarget) Fields() []*vom.RawBytes {
	return targ.fields
}

type structSplitFieldsTarget struct {
	targ *structSplitTarget
}

func (ft *structSplitFieldsTarget) StartField(name string) (key, field vdl.Target, _ error) {
	_, index := ft.targ.tt.FieldByName(name)
	rb := new(vom.RawBytes)
	ft.targ.fields[index] = rb
	return nil, rb.MakeVDLTarget(), nil
}

func (ft *structSplitFieldsTarget) FinishField(key, field vdl.Target) error {
	return nil
}

func (ft *structSplitFieldsTarget) ZeroField(name string) (error) {
	key, field, err := ft.StartField(name)
	if err != nil {
		return err
	}
	fld, index := ft.targ.tt.FieldByName(name)
	if index < 0 {
		return vdl.ErrFieldNoExist
	}
	if err := vdl.FromValue(field, vdl.ZeroValue(fld.Type)); err != nil {
		return err
	}
	return ft.FinishField(key, field)
}

func StructSplitTarget() *structSplitTarget {
	return &structSplitTarget{}
}

func JoinRawBytesAsStruct(targ vdl.Target, structType *vdl.Type, fields []*vom.RawBytes) error {
	st, err := targ.StartFields(structType)
	if err != nil {
		return err
	}
	if structType.NumField() != len(fields) {
		return fmt.Errorf("received %d fields, but %v has %d fields", len(fields), structType, structType.NumField())
	}
	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		k, t, err := st.StartField(f.Name)
		if err != nil {
			return err
		}
		if err := fields[i].FillVDLTarget(t, f.Type); err != nil {
			return err
		}
		if err := st.FinishField(k, t); err != nil {
			return err
		}
	}
	return targ.FinishFields(st)
}
