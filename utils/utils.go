package utils

import (
	"math"
	"reflect"
	"unsafe"
)

// CV computes the coefficient of variation of the set of values in `s`.
//
// If your set has all the values you're interested in, pass a sample=false and
// population standard deviation will be used, otherwise if your set does not
// represent the entire picture, use sample=true.
func CV(s []int, sample bool) float64 {
	t, n64 := 0, float64(len(s))
	for _, n := range s {
		t += n
	}
	mean := float64(t) / n64

	var t2 float64 = 0
	for _, in := range s {
		in64 := float64(in)
		t2 += (in64 - mean) * (in64 - mean)
	}
	if sample {
		n64--
	}
	stddev := math.Sqrt(t2 / n64)

	return stddev / mean * 100
}

func StrPtr(s string) *string {
	return &s
}

func StringToByte(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

// Prepend uses append and copy to make the inverse operation of append,
// returning an slice with src before dst, with as few allocations as possible.
//
// If `cap(dst) >= cap(src) + cap(dst)` prepend does not allocate.
//
// Returns `dst` with the new length, so use it with `a = prepend(a, b)`.
// Otherwise with just `prepend(a, b)` a will have the old length.
func Prepend(dst []byte, src []byte) []byte {
	l := len(src)
	// Add as many empty 0 to dst as src len
	for i := 0; i < l; i++ {
		// If there is spare capacity append extends dst length, otherwise it
		// allocates
		dst = append(dst, 0)
	}
	// copy dst to the second half. Note: dst[:] = dst[:len(dst)]
	copy(dst[l:], dst[:])
	// copy src to the first half
	copy(dst[:l], src)
	// return dst with the new length
	return dst
}
