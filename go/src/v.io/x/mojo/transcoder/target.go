// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"encoding/binary"
	"fmt"
	"math"

	"v.io/v23/vdl"
)

type target struct {
	currentBitOffset uint8
	current          bytesRef
}

func (t target) allocator() *allocator {
	return t.current.allocator
}

func (t target) FromBool(src bool, tt *vdl.Type) error {
	if src {
		t.current.Bytes()[0] |= 1 << t.currentBitOffset
	}
	// shouldn't have to do anything for false
	return nil
}
func (t target) FromUint(src uint64, tt *vdl.Type) error {
	switch tt.Kind() {
	case vdl.Byte:
		t.current.Bytes()[0] = byte(src)
	case vdl.Uint16:
		binary.LittleEndian.PutUint16(t.current.Bytes(), uint16(src))
	case vdl.Uint32:
		binary.LittleEndian.PutUint32(t.current.Bytes(), uint32(src))
	case vdl.Uint64:
		binary.LittleEndian.PutUint64(t.current.Bytes(), src)
	default:
		return fmt.Errorf("invalid FromUint(%v, %v)", src, tt)
	}
	return nil
}
func (t target) FromInt(src int64, tt *vdl.Type) error {
	switch tt.Kind() {
	case vdl.Int16:
		binary.LittleEndian.PutUint16(t.current.Bytes(), uint16(src))
	case vdl.Int32:
		binary.LittleEndian.PutUint32(t.current.Bytes(), uint32(src))
	case vdl.Int64:
		binary.LittleEndian.PutUint64(t.current.Bytes(), uint64(src))
	default:
		return fmt.Errorf("invalid FromInt(%v, %v)", src, tt)
	}
	return nil
}
func (t target) FromFloat(src float64, tt *vdl.Type) error {
	switch tt.Kind() {
	case vdl.Float32:
		binary.LittleEndian.PutUint32(t.current.Bytes(), math.Float32bits(float32(src)))
	case vdl.Float64:
		binary.LittleEndian.PutUint64(t.current.Bytes(), math.Float64bits(src))
	default:
		return fmt.Errorf("invalid FromFloat(%v, %v)", src, tt)
	}
	return nil
}
func (t target) FromComplex(src complex128, tt *vdl.Type) error {
	panic("UNIMPLEMENTED")
}
func (t target) writeBytes(src []byte) {
	block := t.allocator().Allocate(uint32(len(src)), uint32(len(src)))
	t.writePointer(block)
	copy(block.Bytes(), src)
}
func (t target) FromBytes(src []byte, tt *vdl.Type) error {
	t.writeBytes(src)
	return nil
}
func (t target) FromString(src string, tt *vdl.Type) error {
	t.writeBytes([]byte(src))
	return nil
}
func (t target) FromEnumLabel(src string, tt *vdl.Type) error {
	// enums in mojo are treated as an int32 on the wire (but have gaps in their values).
	// This implementation assumes that we will use generated VDL values and not have gaps.
	index := tt.EnumIndex(src)
	binary.LittleEndian.PutUint32(t.current.Bytes(), uint32(index))
	return nil
}
func (t target) FromTypeObject(src *vdl.Type) error {
	panic("UNIMPLEMENTED")

}
func (t target) FromNil(tt *vdl.Type) error {
	switch tt.Kind() {
	case vdl.Optional:
		elemType := tt.Elem()
		switch elemType.Kind() {
		case vdl.Union, vdl.Struct: // Array? String? Bytes? List? Set?
			// Note: for union, this zeros 16 bytes, but for others it does just 8.
			zeroBytes(t.current.Bytes())
		default:
			panic(fmt.Sprintf("Vdl type %v cannot be optional", tt))
		}
	case vdl.Any:
		panic("Any rep not yet determined")
	default:
		panic("Type cannot be nil")
	}
	return nil
}
func zeroBytes(dat []byte) {
	copy(dat, make([]byte, len(dat)))
}

func (t target) StartList(tt *vdl.Type, len int) (vdl.ListTarget, error) {
	if tt.Kind() == vdl.Optional {
		tt = tt.Elem()
	}
	bitsNeeded := baseTypeSizeBits(tt.Elem()) * uint32(len)
	block := t.allocator().Allocate((bitsNeeded+7)/8, uint32(len))
	t.writePointer(block)
	if tt.Elem().Kind() == vdl.Bool {
		return &bitListTarget{
			block: block,
		}, nil
	} else {
		return &listTarget{
			incrementSize: baseTypeSizeBits(tt.Elem()) / 8,
			block:         block,
		}, nil
	}
}
func (t target) FinishList(x vdl.ListTarget) error {
	return nil
}

// TODO(bprosnitz) This uses list, should we use map instead?
func (t target) StartSet(tt *vdl.Type, len int) (vdl.SetTarget, error) {
	if tt.Kind() == vdl.Optional {
		tt = tt.Elem()
	}
	bitsNeeded := baseTypeSizeBits(tt.Key()) * uint32(len)
	block := t.allocator().Allocate((bitsNeeded+7)/8, uint32(len))
	t.writePointer(block)
	if tt.Key().Kind() == vdl.Bool {
		return &bitListTarget{
			block: block,
		}, nil
	} else {
		return &listTarget{
			incrementSize: baseTypeSizeBits(tt.Key()) / 8,
			block:         block,
		}, nil
	}
}
func (t target) FinishSet(x vdl.SetTarget) error {
	return nil
}
func (t target) StartMap(tt *vdl.Type, len int) (vdl.MapTarget, error) {
	if tt.Kind() == vdl.Optional {
		tt = tt.Elem()
	}
	pointerBlock := t.allocator().Allocate(16, 0)
	t.writePointer(pointerBlock)

	st := target{
		current: pointerBlock.Slice(0, 8),
	}
	setTarget, err := st.StartSet(vdl.SetType(tt.Key()), len)
	if err != nil {
		return nil, err
	}
	lt := target{
		current: pointerBlock.Slice(8, 16),
	}
	listTarget, err := lt.StartList(vdl.ListType(tt.Elem()), len)
	if err != nil {
		return nil, err
	}
	return &mapTarget{
		setTarget:  setTarget,
		listTarget: listTarget,
	}, nil
}
func (t target) FinishMap(x vdl.MapTarget) error {
	return nil
}
func (t target) StartFields(tt *vdl.Type) (vdl.FieldsTarget, error) {
	if tt.Kind() == vdl.Optional {
		tt = tt.Elem()
	}
	if tt.Kind() == vdl.Union {
		return unionFieldsTarget{
			vdlType: tt,
			block:   t.current,
		}, nil
	}
	block := t.allocator().Allocate(neededStructAllocationSize(tt), 0)
	t.writePointer(block)
	return fieldsTarget{
			vdlType: tt,
			block:   block,
		},
		nil
}
func (t target) FinishFields(x vdl.FieldsTarget) error {
	return nil
}

func (t target) writePointer(alloc bytesRef) {
	offset := alloc.AsPointer(t.current)
	binary.LittleEndian.PutUint64(t.current.Bytes(), uint64(offset))
}
