//go:build purego || !(amd64 || arm64)

package rlnc

func supportedKernels() []kernel {
	return []kernel{kernelGeneric}
}

func mulKernel(kernel, []byte, []byte, byte) int {
	return 0
}

func xorKernel(kernel, []byte, []byte, byte) int {
	return 0
}
