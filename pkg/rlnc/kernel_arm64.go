//go:build !purego

package rlnc

// supportedKernels returns NEON on every arm64 CPU, since Advanced SIMD is part
// of the base architecture.
func supportedKernels() []kernel {
	return []kernel{kernelNEON, kernelGeneric}
}

// mulKernel runs k over the longest prefix of dst it handles and returns that
// prefix's length.
func mulKernel(k kernel, dst, src []byte, factor byte) int {
	if k != kernelNEON {
		return 0
	}

	n := len(dst) &^ 15
	neonMul(&nibbleTables[factor], dst[:n], src[:n])

	return n
}

// xorKernel is mulKernel for xorSlice.
func xorKernel(k kernel, dst, src []byte, factor byte) int {
	if k != kernelNEON {
		return 0
	}

	n := len(dst) &^ 15
	neonXor(&nibbleTables[factor], dst[:n], src[:n])

	return n
}

// neonMul and neonXor take a dst whose length is a nonzero multiple of 16, and
// a src at least as long.
//
//go:noescape
func neonMul(tables *[32]byte, dst, src []byte)

//go:noescape
func neonXor(tables *[32]byte, dst, src []byte)
