//go:build !purego

package rlnc

import "github.com/klauspost/cpuid/v2"

// supportedKernels returns the kernels this CPU and OS support, fastest first.
// cpuid also checks that the OS saves the vector registers each one uses.
func supportedKernels() []kernel {
	supported := make([]kernel, 0, 4)

	if cpuid.CPU.Supports(cpuid.AVX512F, cpuid.AVX512BW, cpuid.GFNI) {
		supported = append(supported, kernelGFNIAVX512)
	}
	if cpuid.CPU.Supports(cpuid.AVX2, cpuid.GFNI) {
		supported = append(supported, kernelGFNI)
	}
	if cpuid.CPU.Supports(cpuid.AVX2) {
		supported = append(supported, kernelAVX2)
	}

	return append(supported, kernelGeneric)
}

// mulKernel runs k over the longest prefix of dst it handles and returns that
// prefix's length.
func mulKernel(k kernel, dst, src []byte, factor byte) int {
	switch k {
	case kernelGFNIAVX512:
		gfniMulAVX512(gfniMatrices[factor], dst, src)
		return len(dst)

	case kernelGFNI:
		n := len(dst) &^ 15
		gfniMulAVX2(gfniMatrices[factor], dst[:n], src[:n])
		return n

	case kernelAVX2:
		n := len(dst) &^ 15
		avx2Mul(&nibbleTables[factor], dst[:n], src[:n])
		return n

	default:
		return 0
	}
}

// xorKernel is mulKernel for xorSlice.
func xorKernel(k kernel, dst, src []byte, factor byte) int {
	switch k {
	case kernelGFNIAVX512:
		gfniXorAVX512(gfniMatrices[factor], dst, src)
		return len(dst)

	case kernelGFNI:
		n := len(dst) &^ 15
		gfniXorAVX2(gfniMatrices[factor], dst[:n], src[:n])
		return n

	case kernelAVX2:
		n := len(dst) &^ 15
		avx2Xor(&nibbleTables[factor], dst[:n], src[:n])
		return n

	default:
		return 0
	}
}

// The AVX-512 kernels take a dst of any nonzero length. The others take a dst
// whose length is a nonzero multiple of 16. src is at least as long as dst.

//go:noescape
func gfniMulAVX512(matrix uint64, dst, src []byte)

//go:noescape
func gfniXorAVX512(matrix uint64, dst, src []byte)

//go:noescape
func gfniMulAVX2(matrix uint64, dst, src []byte)

//go:noescape
func gfniXorAVX2(matrix uint64, dst, src []byte)

//go:noescape
func avx2Mul(tables *[32]byte, dst, src []byte)

//go:noescape
func avx2Xor(tables *[32]byte, dst, src []byte)
