// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"encoding/binary"

	"v.io/v23/vdl"
)

// for struct
type fieldsTarget struct {
	vdlType *vdl.Type
	block   bytesRef
	layout  structLayout
}

func (fe fieldsTarget) StartField(name string) (key, field vdl.Target, _ error) {
	fieldType, fieldIndex := fe.vdlType.FieldByName(name)
	byteOffset, bitOffset := fe.layout.MojoOffsetsFromVdlIndex(fieldIndex)

	numBits := baseTypeSizeBits(fieldType.Type)
	refSize := (numBits + 7) / 8
	newRef := fe.block.Slice(byteOffset, byteOffset+refSize)
	return nil, target{
		currentBitOffset: bitOffset,
		current:          newRef,
	}, nil
}
func (fieldsTarget) FinishField(key, field vdl.Target) error {
	return nil
}

type unionFieldsTarget struct {
	vdlType *vdl.Type
	block   bytesRef
}

func (ufe unionFieldsTarget) StartField(name string) (key, field vdl.Target, _ error) {
	fld, index := ufe.vdlType.FieldByName(name)
	binary.LittleEndian.PutUint32(ufe.block.Bytes(), 16)
	binary.LittleEndian.PutUint32(ufe.block.Bytes()[4:], uint32(index))

	valueSlice := ufe.block.Slice(8, 16)
	if fld.Type.Kind() == vdl.Union {
		// nested union, create a pointer to the body
		nestedUnionBlock := ufe.block.allocator.Allocate(8, 0)
		offset := nestedUnionBlock.AsPointer(valueSlice)
		binary.LittleEndian.PutUint64(valueSlice.Bytes(), uint64(offset))
		return nil, target{
			currentBitOffset: 0,
			current:          nestedUnionBlock.SignedSlice(-8, 8),
		}, nil
	}
	return nil, target{
		currentBitOffset: 0,
		current:          valueSlice,
	}, nil
}
func (unionFieldsTarget) FinishField(key, field vdl.Target) error {
	return nil
}

// doubles as set target
type listTarget struct {
	incrementSize uint32
	block         bytesRef
	nextPosition  uint32
}

func (lt *listTarget) StartElem(index int) (elem vdl.Target, _ error) {
	// TODO(bprosnitz) Index is ignored -- we should probably remove this from Target.
	sliceBlock := lt.block.Slice(lt.nextPosition, lt.nextPosition+lt.incrementSize)
	lt.nextPosition += lt.incrementSize
	return target{
		current: sliceBlock,
	}, nil
}
func (lt *listTarget) StartKey() (key vdl.Target, _ error) {
	return lt.StartElem(0)
}
func (listTarget) FinishElem(elem vdl.Target) error {
	return nil
}
func (listTarget) FinishKey(key vdl.Target) error {
	return nil
}

type bitListTarget struct {
	block           bytesRef
	nextBitPosition uint32
}

func (blt *bitListTarget) StartElem(index int) (elem vdl.Target, _ error) {
	bitPos := blt.nextBitPosition
	blt.nextBitPosition++
	byteIndex := bitPos / 8
	sliceBlock := blt.block.Slice(byteIndex, byteIndex+1)
	return target{
		currentBitOffset: uint8(bitPos % 8),
		current:          sliceBlock,
	}, nil
}
func (blt *bitListTarget) StartKey() (key vdl.Target, _ error) {
	return blt.StartElem(0)
}
func (bitListTarget) FinishElem(elem vdl.Target) error {
	return nil
}
func (bitListTarget) FinishKey(key vdl.Target) error {
	return nil
}

type mapTarget struct {
	keys             vdl.SetTarget
	valuePlaceholder vdl.Target
	valueType        *vdl.Type
	cachedValues     []*vdl.Value
}

func (mt *mapTarget) StartKey() (key vdl.Target, _ error) {
	return mt.keys.StartKey()
}
func (mt *mapTarget) FinishKeyStartField(key vdl.Target) (field vdl.Target, err error) {
	val := vdl.ZeroValue(mt.valueType)
	field, err = vdl.ValueTarget(val)
	mt.cachedValues = append(mt.cachedValues, val)
	return
}
func (mapTarget) FinishField(key, field vdl.Target) error {
	return nil
}
