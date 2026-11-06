package rlnc

// GFMultiply multiplies two elements of the RLNC encoder's GF(256) field, which uses
// the primitive polynomial 0x11d.
func GFMultiply(left, right byte) byte {
	return mulTable[left][right]
}

// multiply is the bitwise GF(256) product that mulTable is built from.
func multiply(left, right byte) byte {
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

// GFInverse returns the multiplicative inverse of value in GF(256). Zero has no
// inverse and returns zero.
func GFInverse(value byte) byte {
	if value == 0 {
		return 0
	}

	// a^254 is a^-1 in GF(256).
	result := byte(1)
	base := value
	for exponent := 254; exponent > 0; exponent >>= 1 {
		if exponent&1 != 0 {
			result = GFMultiply(result, base)
		}

		base = GFMultiply(base, base)
	}

	return result
}

// Scale multiplies every element of row by factor in place.
func Scale(row []byte, factor byte) {
	if factor == 1 {
		return
	}

	products := &mulTable[factor]
	for index, value := range row {
		row[index] = products[value]
	}
}

// XorScaled adds factor*src to dst in place. Addition in GF(256) is XOR. src
// must be at least as long as dst. dst and src may overlap, and are then
// processed in index order.
func XorScaled(dst, src []byte, factor byte) {
	if factor == 0 {
		return
	}

	products := &mulTable[factor]
	for index := range dst {
		dst[index] ^= products[src[index]]
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

// mulTable holds every GF(256) product, so a multiply is one lookup instead of
// a bitwise loop. It is 64 KiB, and a row for one factor fits in L1.
var mulTable = func() (table [256][256]byte) {
	for left := range 256 {
		for right := range 256 {
			table[left][right] = multiply(byte(left), byte(right))
		}
	}

	return table
}()
