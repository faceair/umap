package umap

import (
	"math/bits"
	"unsafe"
)

const (
	bucketCnt = 8
)

type sliceHeader struct {
	Data unsafe.Pointer
}

func makeUint64BucketArray(size int) unsafe.Pointer {
	x := make([]bmapuint64, size)
	for i := range x {
		*(*uint64)(unsafe.Pointer(&x[i].tophash)) = allEmpty
	}
	return (*sliceHeader)(unsafe.Pointer(&x)).Data
}

func (b *bmapuint64) MatchEmptyOrDeleted() bitmask64 {
	// The high bit is set for both empty slot and deleted slot.
	ctrl := littleEndianBytesToUint64(b.tophash)
	return bitmask64(msbs & ctrl)
}

func (b *bmapuint64) MatchEmpty() bitmask64 {
	// Same as b.MatchTopHash(emptySlot), but faster.
	//
	// The high bit is set for both empty slot and deleted slot.
	// (ctrl & emptyOrDeletedMask) get all empty or deleted slots.
	// (ctrl << 1) clears the high bit for deletedSlot.
	// ANDing them we can get all the empty slots.
	ctrl := littleEndianBytesToUint64(b.tophash)
	return bitmask64((ctrl << 1) & ctrl & msbs)
}

func matchTopHash(tophash [bucketCnt]uint8, top uint8) bitmask64 {
	ctrl := littleEndianBytesToUint64(tophash)
	cmp := ctrl ^ (lsbs * uint64(top))
	return bitmask64((cmp - lsbs) & ^cmp & msbs)
}

func (b *bmapuint64) PrepareSameSizeGrow() {
	// Convert Deleted to Empty and Full to Deleted.
	ctrl := littleEndianBytesToUint64(b.tophash)
	full := ^ctrl & msbs
	full = ^full + (full >> 7)
	b.tophash = littleEndianUint64ToBytes(full)
}

func (b bitmask64) AnyMatch() bool {
	return b != 0
}

func (b *bitmask64) NextMatch() uint {
	return uint(bits.TrailingZeros64(uint64(*b)) / bucketCnt)
}

func (b *bitmask64) RemoveLowestBit() {
	*b = *b & (*b - 1)
}

func littleEndianBytesToUint64(v [8]uint8) uint64 {
	return uint64(v[0]) | uint64(v[1])<<8 | uint64(v[2])<<16 | uint64(v[3])<<24 | uint64(v[4])<<32 | uint64(v[5])<<40 | uint64(v[6])<<48 | uint64(v[7])<<56
}

func littleEndianUint64ToBytes(v uint64) [8]uint8 {
	return [8]uint8{uint8(v), uint8(v >> 8), uint8(v >> 16), uint8(v >> 24), uint8(v >> 32), uint8(v >> 40), uint8(v >> 48), uint8(v >> 56)}
}
