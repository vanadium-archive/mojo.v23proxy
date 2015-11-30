// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"fmt"

	"v.io/v23/vdl"
)

// an array where the index corresponds to the bit index in the allocation
// and the value is the tag (here representing vdlIndex+1)
type structBitAllocation []int

// allocateStructBits performs the naive allocation of fields in the struct,
// literally laying out all bits in an array - each associated with a tag
// and finding the first spot where a block of the given size fits (given
// alignment constraints)
func allocateStructBits(a structBitAllocation, tag, size int) structBitAllocation {
	lenActive := 0
	// Scan the given array for a run of |size| empty locations.
	// If found, fill that section with tag.
	for i, v := range a {
		if v == 0 {
			lenActive++
		} else {
			lenActive = 0
		}

		if i%size == size-1 && lenActive >= size {
			for j := i - size + 1; j <= i; j++ {
				a[j] = tag
			}
			return a
		}
	}

	// If there isn't a sufficiently large empty location, allocate a new aligned block.
	paddingAmt := size - len(a)%size
	if paddingAmt == size {
		paddingAmt = 0
	}
	for i := 0; i < paddingAmt; i++ {
		a = append(a, 0)
	}
	// Now len(a) % size == 0
	for i := 0; i < size; i++ {
		a = append(a, tag)
	}

	return a
}

type structLayoutField struct {
	vdlStructIndex int    // the index in the vdl value of this field
	byteOffset     uint32 // byte offset from the beginning of the mojom byte array
	bitOffset      uint8  // bit offset [0,8)
}

// computeStructLayout computes a representation of the fields in a struct, as
// a list ordered by mojom byte field order.
func computeStructLayout(t *vdl.Type) (layout structLayout) {
	a := structBitAllocation{}

	for i := 0; i < t.NumField(); i++ {
		bits := baseTypeSizeBits(t.Field(i).Type)
		a = allocateStructBits(a, i+1, int(bits))
	}

	lastVal := 0
	for i, v := range a {
		if v != lastVal && v != 0 {
			layout = append(layout, structLayoutField{
				vdlStructIndex: v - 1,
				byteOffset:     uint32(i / 8),
				bitOffset:      uint8(i % 8),
			})
			lastVal = v
		}
	}

	return
}

type structLayout []structLayoutField

func (s structLayout) MojoOffsetsFromVdlIndex(vdlIndex int) (byteOffset uint32, bitOffset uint8) {
	for _, alloc := range s {
		if alloc.vdlStructIndex == vdlIndex {
			return alloc.byteOffset, alloc.bitOffset
		}
	}
	panic(fmt.Sprintf("unknown vdl index %d (layout %v) -- this should never happen", vdlIndex, s))
}
