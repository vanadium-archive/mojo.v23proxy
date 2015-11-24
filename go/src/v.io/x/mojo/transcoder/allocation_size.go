// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import "v.io/v23/vdl"

func neededStructAllocationSize(vt *vdl.Type) uint32 {
	var totalBits uint32
	for fi := 0; fi < vt.NumField(); fi++ {
		field := vt.Field(fi)
		totalBits += baseTypeSizeBits(field.Type)
	}
	return roundBitsTo64Alignment(totalBits)
}

func baseTypeSizeBits(vt *vdl.Type) uint32 {
	switch vt.Kind() {
	case vdl.Bool:
		return 1
	case vdl.Byte, vdl.Int8:
		return 8
	case vdl.Uint16, vdl.Int16:
		return 16
	case vdl.Uint32, vdl.Int32, vdl.Float32, vdl.Enum:
		return 32
	case vdl.Union:
		return 128 // Header + value / pointer to inner union
	default: // Either Uint64, Int64, Float64 or pointer.
		return 64
	}
}

// Round up to the nearest 8 byte length.
func roundBitsTo64Alignment(numBits uint32) uint32 {
	if numBits%64 == 0 {
		return numBits / 8
	}
	return (numBits + (64 - numBits%64)) / 8
}
