package rlnc

// Mul multiplies two elements of the RLNC encoder's GF(256) field, which uses
// the primitive polynomial 0x11d.
func Mul(left, right byte) byte {
	var product byte
	for right != 0 {
		if right&1 != 0 {
			product ^= left
		}

		carry := left&0x80 != 0
		left <<= 1

		if carry {
			left ^= 0x1d
		}

		right >>= 1
	}

	return product
}

// Inverse returns the multiplicative inverse of value in GF(256). Zero has no
// inverse and returns zero.
func Inverse(value byte) byte {
	if value == 0 {
		return 0
	}

	// a^254 is a^-1 in GF(256).
	result := byte(1)
	base := value
	for exponent := 254; exponent > 0; exponent >>= 1 {
		if exponent&1 != 0 {
			result = Mul(result, base)
		}

		base = Mul(base, base)
	}

	return result
}

// Scale multiplies every element of row by factor in place.
func Scale(row []byte, factor byte) {
	for index := range row {
		row[index] = Mul(factor, row[index])
	}
}

// XorScaled adds factor*src to dst in place. Addition in GF(256) is XOR. src
// must be at least as long as dst.
func XorScaled(dst, src []byte, factor byte) {
	for index := range dst {
		dst[index] ^= Mul(factor, src[index])
	}
}

// LeadingColumn returns the index of row's first nonzero element, or -1 if row
// is all zeros.
func LeadingColumn(row []byte) int {
	for index, value := range row {
		if value != 0 {
			return index
		}
	}

	return -1
}
