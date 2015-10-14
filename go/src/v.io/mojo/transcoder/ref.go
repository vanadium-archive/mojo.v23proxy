package mojovdl

type bytesRef struct {
	allocator            *allocator
	startIndex, endIndex uint32
}

func (b bytesRef) Slice(low, high uint32) bytesRef {
	return bytesRef{
		allocator:  b.allocator,
		startIndex: b.startIndex + low,
		endIndex:   b.startIndex + high,
	}
}

// SignedSlice allows going backwards in the slice (temporary hack to include header in slice for unions)
func (b bytesRef) SignedSlice(low, high int) bytesRef {
	return bytesRef{
		allocator:  b.allocator,
		startIndex: uint32(int(b.startIndex) + low),
		endIndex:   uint32(int(b.startIndex) + high),
	}
}

func (b bytesRef) Bytes() []byte {
	return b.allocator.buf[b.startIndex:b.endIndex]
}

func (b bytesRef) AsPointer(fromRefPos bytesRef) uint32 {
	offset := b.startIndex - fromRefPos.startIndex
	if offset <= 0 {
		panic("invalid non-positive offset for pointer")
	}
	return offset - HEADER_SIZE
}
