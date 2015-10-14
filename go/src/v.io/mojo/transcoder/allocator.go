package mojovdl

import "encoding/binary"

const HEADER_SIZE uint32 = 8

// TODO(bprosnitz) maybe make an interface to support decode too
type allocator struct {
	// Buffer containing encoded data.
	buf []byte

	// Index of the first unclaimed byte in buf.
	end uint32
}

func (a *allocator) makeRoom(size uint32) {
	totalNeeded := a.end + size

	allocationSize := uint32(1)
	for allocationSize < totalNeeded {
		allocationSize *= 2
	}

	if allocationSize != uint32(len(a.buf)) {
		oldBuf := a.buf
		a.buf = make([]byte, allocationSize)
		copy(a.buf, oldBuf)
	}
}

func (a *allocator) allocateBlock(size uint32, numElems uint32) (startIndex, endIndex uint32) {
	size_with_header := size + HEADER_SIZE
	size_with_header_rounded := size_with_header
	if size_with_header%8 != 0 {
		size_with_header_rounded = size_with_header + (8 - (size_with_header % 8))
	}

	a.makeRoom(size_with_header_rounded)
	binary.LittleEndian.PutUint32(a.buf[a.end:a.end+4], size_with_header)
	binary.LittleEndian.PutUint32(a.buf[a.end+4:a.end+8], numElems)

	prevEnd := a.end
	start := prevEnd + HEADER_SIZE
	end := prevEnd + size_with_header
	a.end = prevEnd + size_with_header_rounded
	return start, end
}

func (a *allocator) Allocate(size uint32, numElems uint32) bytesRef {
	begin, end := a.allocateBlock(size, numElems)
	ref := bytesRef{
		allocator:  a,
		startIndex: begin,
		endIndex:   end,
		// zeros
	}
	return ref
}

func (a *allocator) AllocationFromPointer(absoluteIndex uint32) bytesRef {
	headerPos := absoluteIndex - HEADER_SIZE
	size := binary.LittleEndian.Uint32(a.buf[headerPos : headerPos+4])
	return bytesRef{
		allocator:  a,
		startIndex: absoluteIndex,
		endIndex:   absoluteIndex + size,
	}
}

func (a *allocator) AllocatedBytes() []byte {
	if a.buf == nil {
		return []byte{}
	}
	return a.buf[:a.end]
}
