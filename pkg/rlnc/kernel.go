package rlnc

import "unsafe"

// kernel names an implementation of the GF(256) slice operations behind Scale
// and XorScaled.
type kernel uint8

const (
	// kernelGeneric looks every byte up in mulTable. It runs on every platform
	// and finishes the tail that a vector kernel leaves.
	kernelGeneric kernel = iota
	// kernelNEON splits each byte into nibbles and looks both up with arm64
	// TBL, 16 bytes per instruction.
	kernelNEON
	// kernelAVX2 is the nibble method with amd64 VPSHUFB, 32 bytes at a time.
	kernelAVX2
	// kernelGFNI applies multiplication as an 8×8 bit matrix with
	// VGF2P8AFFINEQB on 32 bytes at a time.
	kernelGFNI
	// kernelGFNIAVX512 is kernelGFNI on 64 bytes at a time. Masked loads and
	// stores finish the tail, so it leaves nothing for kernelGeneric.
	kernelGFNIAVX512
)

// minKernelLen is the shortest slice handed to a vector kernel. Shorter ones
// cost less as table lookups than as a call and a partial vector.
const minKernelLen = 16

func (k kernel) String() string {
	switch k {
	case kernelGeneric:
		return "generic"
	case kernelNEON:
		return "neon"
	case kernelAVX2:
		return "avx2"
	case kernelGFNI:
		return "gfni"
	case kernelGFNIAVX512:
		return "gfni-avx512"
	default:
		return "unknown"
	}
}

var (
	// kernels lists the kernels this CPU supports, fastest first. The last is
	// always kernelGeneric.
	kernels = supportedKernels()

	// activeKernel is the kernel Scale and XorScaled use. Tests switch it to
	// cover every kernel in kernels.
	activeKernel = kernels[0]
)

// mulSlice sets dst[i] = factor·src[i] for every index of dst. src must be at
// least as long as dst.
func mulSlice(dst, src []byte, factor byte) {
	done := 0
	if len(dst) >= minKernelLen {
		done = mulKernel(activeKernel, dst, src, factor)
	}

	products := &mulTable[factor]
	for index := done; index < len(dst); index++ {
		dst[index] = products[src[index]]
	}
}

// xorSlice adds factor·src[i] to dst[i] for every index of dst. src must be at
// least as long as dst.
func xorSlice(dst, src []byte, factor byte) {
	done := 0
	if len(dst) >= minKernelLen && !overlapsBehind(dst, src) {
		done = xorKernel(activeKernel, dst, src, factor)
	}

	products := &mulTable[factor]
	for index := done; index < len(dst); index++ {
		dst[index] ^= products[src[index]]
	}
}

// overlapsBehind reports whether dst starts inside src after its first byte.
// A vector kernel reads a block of src before it writes the same block of dst,
// so it would read src bytes that index order has already updated through dst.
func overlapsBehind(dst, src []byte) bool {
	if len(src) == 0 {
		return false
	}

	dstStart := uintptr(unsafe.Pointer(unsafe.SliceData(dst)))
	srcStart := uintptr(unsafe.Pointer(unsafe.SliceData(src)))

	return dstStart > srcStart && dstStart-srcStart < uintptr(len(src))
}

// nibbleTables[f] holds f·x for every low nibble x in its first 16 bytes, and
// f·(x<<4) for every high nibble x in its last 16. Multiplication distributes
// over XOR, so f·b is the XOR of the lookups for b's two nibbles. Shuffle
// kernels do both lookups for a whole vector at once.
var nibbleTables = func() (tables [256][32]byte) {
	for factor := range 256 {
		for nibble := range 16 {
			tables[factor][nibble] = mulTable[factor][nibble]
			tables[factor][16+nibble] = mulTable[factor][nibble<<4]
		}
	}

	return tables
}()

// gfniMatrices[f] is multiplication by f as an 8×8 bit matrix, in the layout
// GF2P8AFFINEQB takes. Byte 7-i of the matrix selects the input bits whose
// parity is output bit i. Multiplication by a constant is linear over GF(2),
// so any field polynomial works, unlike GF2P8MULB, which is fixed to 0x11b.
var gfniMatrices = func() (matrices [256]uint64) {
	for factor := range 256 {
		for output := range 8 {
			var row uint64
			for input := range 8 {
				if mulTable[factor][1<<input]&(1<<output) != 0 {
					row |= 1 << input
				}
			}

			matrices[factor] |= row << (8 * (7 - output))
		}
	}

	return matrices
}()
