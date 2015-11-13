// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transcoder

import (
	"fmt"

	"v.io/v23/vdl"
)

func offsetInStruct(index int, structVdlType *vdl.Type) (byteOffset uint32, bitOffset uint8) {
	if index >= structVdlType.NumField() {
		panic(fmt.Sprintf("index %v out of bounds for type %v", index, structVdlType))
	}

	offsetComputation := &structOffsetComputation{}
	for i := 0; i < index; i++ {
		bits := baseTypeSizeBits(structVdlType.Field(i).Type)
		switch bits {
		case 1:
			offsetComputation.next1()
		case 8:
			offsetComputation.next8()
		case 16:
			offsetComputation.next16()
		case 32:
			offsetComputation.next32()
		case 64:
			offsetComputation.next64()
		case 128:
			offsetComputation.next64()
			offsetComputation.next64()
		default:
			panic("unknown bit size")
		}
	}

	switch baseTypeSizeBits(structVdlType.Field(index).Type) {
	case 1:
		return offsetComputation.index1, offsetComputation.bitOffset1
	case 8:
		return offsetComputation.index8, 0
	case 16:
		return offsetComputation.index16, 0
	case 32:
		return offsetComputation.index32, 0
	case 64, 128:
		return offsetComputation.index64, 0

	default:
		panic("unknown bit size")
	}
}

type structOffsetComputation struct {
	index8, index16, index32, index64 uint32
	index1                            uint32
	bitOffset1                        uint8
}

func (sa *structOffsetComputation) next1() {
	newBitOffset1 := (sa.bitOffset1 + 1) % 8
	var nextIndex uint32
	if sa.index1 != sa.index8 {
		nextIndex = sa.index8
	} else {
		nextIndex = sa.index1 + 1
	}
	if sa.bitOffset1 == 0 {
		// bit offset == 0 means fetch a new byte
		if sa.index8 == sa.index1 {
			sa.index8 = nextIndex
		}
		if sa.index16 == sa.index1 {
			if nextIndex&0x1 == 0 {
				sa.index16 = nextIndex
			} else {
				sa.index16 = nextIndex + 1
			}
		}
		if sa.index32 == sa.index1 {
			if nextIndex&0x3 == 0 {
				sa.index32 = nextIndex
			} else {
				sa.index32 = nextIndex + (4 - (nextIndex & 0x3))
			}
		}
		if sa.index64 == sa.index1 {
			if nextIndex&0x7 == 0 {
				sa.index64 = nextIndex
			} else {
				sa.index64 = nextIndex + (8 - (nextIndex & 0x7))
			}
		}
	}

	if sa.bitOffset1 == 7 {
		sa.index1 = nextIndex
	}
	sa.bitOffset1 = newBitOffset1
}
func (sa *structOffsetComputation) next8() {
	var newIndex8 uint32
	if sa.index8 != sa.index16 {
		newIndex8 = sa.index16
	} else {
		newIndex8 = sa.index8 + 1
	}

	if sa.index1 == sa.index8 && sa.bitOffset1 == 0 {
		sa.index1 = newIndex8
	}
	if sa.index16 == sa.index8 {
		if newIndex8&0x1 == 0 {
			sa.index16 = newIndex8
		} else {
			sa.index16 = newIndex8 + 1
		}
	}
	if sa.index32 == sa.index8 {
		if newIndex8&0x3 == 0 {
			sa.index32 = newIndex8
		} else {
			sa.index32 = newIndex8 + (4 - (newIndex8 & 0x3))
		}
	}
	if sa.index64 == sa.index8 {
		if newIndex8&0x7 == 0 {
			sa.index64 = newIndex8
		} else {
			sa.index64 = newIndex8 + (8 - (newIndex8 & 0x7))
		}
	}
	sa.index8 = newIndex8
}
func (sa *structOffsetComputation) next16() {
	var newIndex16 uint32
	if sa.index16 != sa.index32 {
		newIndex16 = sa.index32
	} else {
		newIndex16 = sa.index16 + 2
	}
	if sa.index1 == sa.index16 && sa.bitOffset1 == 0 {
		sa.index1 = newIndex16
	}
	if sa.index8 == sa.index16 {
		sa.index8 = newIndex16
	}
	if sa.index32 == sa.index16 {
		if newIndex16&0x3 == 0 {
			sa.index32 = newIndex16
		} else {
			sa.index32 = newIndex16 + (4 - (newIndex16 & 0x3))
		}
	}
	if sa.index64 == sa.index16 {
		if newIndex16&0x7 == 0 {
			sa.index64 = newIndex16
		} else {
			sa.index64 = newIndex16 + (8 - (newIndex16 & 0x7))
		}
	}
	sa.index16 = newIndex16
}
func (sa *structOffsetComputation) next32() {
	var newIndex32 uint32
	if sa.index32 != sa.index64 {
		newIndex32 = sa.index64
	} else {
		newIndex32 = sa.index32 + 4
	}
	if sa.index1 == sa.index32 && sa.bitOffset1 == 0 {
		sa.index1 = newIndex32
	}
	if sa.index8 == sa.index32 {
		sa.index8 = newIndex32
	}
	if sa.index16 == sa.index32 {
		sa.index16 = newIndex32
	}
	if sa.index64 == sa.index32 {
		if newIndex32&0x7 == 0 {
			sa.index64 = newIndex32
		} else {
			sa.index64 = newIndex32 + (8 - (newIndex32 & 0x7))
		}
	}
	sa.index32 = newIndex32
}
func (sa *structOffsetComputation) next64() {
	newIndex64 := sa.index64 + 8

	if sa.index1 == sa.index64 && sa.bitOffset1 == 0 {
		sa.index1 = newIndex64
	}
	if sa.index8 == sa.index64 {
		sa.index8 = newIndex64
	}
	if sa.index16 == sa.index64 {
		sa.index16 = newIndex64
	}
	if sa.index32 == sa.index64 {
		sa.index32 = newIndex64
	}
	sa.index64 = newIndex64
}
