// Package rlnc provides GF(256) arithmetic, basis reduction and sharder constants
// for code that works with RLNC symbols outside rlnc-server.
//
// Scale and XorScaled run on the fastest vector kernel the CPU supports, chosen
// once at startup. On amd64 that is GFNI with AVX-512, GFNI with AVX2, or AVX2,
// and on arm64 it is NEON. Every other platform, and any build with the purego
// tag, uses 64 KiB of lookup tables. All kernels return identical results.
package rlnc
