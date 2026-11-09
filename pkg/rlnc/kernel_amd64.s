//go:build !purego

#include "textflag.h"

// The GFNI kernels multiply with VGF2P8AFFINEQB, which applies an 8×8 bit
// matrix to every byte. The factor's matrix is broadcast to every 64-bit lane.
//
// The AVX2 kernels split each byte b into nibbles, look both up in the
// factor's 16-byte tables with VPSHUFB, and XOR the two products, since f·b is
// f·(b & 15) XOR f·(b & 240).

// func gfniMulAVX512(matrix uint64, dst, src []byte)
TEXT ·gfniMulAVX512(SB), NOSPLIT, $0-56
	VPBROADCASTQ matrix+0(FP), Z0
	MOVQ         dst_base+8(FP), DI
	MOVQ         dst_len+16(FP), CX
	MOVQ         src_base+32(FP), SI
	CMPQ         CX, $64
	JB           mulTail

mul64:
	VMOVDQU64      (SI), Z1
	VGF2P8AFFINEQB $0, Z0, Z1, Z1
	VMOVDQU64      Z1, (DI)
	ADDQ           $64, SI
	ADDQ           $64, DI
	SUBQ           $64, CX
	CMPQ           CX, $64
	JAE            mul64

mulTail:
	// A mask of the CX < 64 remaining bytes. Masked-out bytes are neither read
	// nor written, so the tail never touches memory past the slices.
	TESTQ          CX, CX
	JZ             mulDone
	MOVQ           $1, AX
	SHLQ           CX, AX
	DECQ           AX
	KMOVQ          AX, K1
	VMOVDQU8.Z     (SI), K1, Z1
	VGF2P8AFFINEQB $0, Z0, Z1, Z1
	VMOVDQU8       Z1, K1, (DI)

mulDone:
	VZEROUPPER
	RET

// func gfniXorAVX512(matrix uint64, dst, src []byte)
TEXT ·gfniXorAVX512(SB), NOSPLIT, $0-56
	VPBROADCASTQ matrix+0(FP), Z0
	MOVQ         dst_base+8(FP), DI
	MOVQ         dst_len+16(FP), CX
	MOVQ         src_base+32(FP), SI
	CMPQ         CX, $64
	JB           xorTail

xor64:
	VMOVDQU64      (SI), Z1
	VGF2P8AFFINEQB $0, Z0, Z1, Z1
	VPXORQ         (DI), Z1, Z1
	VMOVDQU64      Z1, (DI)
	ADDQ           $64, SI
	ADDQ           $64, DI
	SUBQ           $64, CX
	CMPQ           CX, $64
	JAE            xor64

xorTail:
	TESTQ          CX, CX
	JZ             xorDone
	MOVQ           $1, AX
	SHLQ           CX, AX
	DECQ           AX
	KMOVQ          AX, K1
	VMOVDQU8.Z     (SI), K1, Z1
	VMOVDQU8.Z     (DI), K1, Z2
	VGF2P8AFFINEQB $0, Z0, Z1, Z1
	VPXORQ         Z2, Z1, Z1
	VMOVDQU8       Z1, K1, (DI)

xorDone:
	VZEROUPPER
	RET

// func gfniMulAVX2(matrix uint64, dst, src []byte)
TEXT ·gfniMulAVX2(SB), NOSPLIT, $0-56
	VPBROADCASTQ matrix+0(FP), Y0
	MOVQ         dst_base+8(FP), DI
	MOVQ         dst_len+16(FP), CX
	MOVQ         src_base+32(FP), SI
	CMPQ         CX, $32
	JB           mul16

mul32:
	VMOVDQU        (SI), Y1
	VGF2P8AFFINEQB $0, Y0, Y1, Y1
	VMOVDQU        Y1, (DI)
	ADDQ           $32, SI
	ADDQ           $32, DI
	SUBQ           $32, CX
	CMPQ           CX, $32
	JAE            mul32

mul16:
	TESTQ          CX, CX
	JZ             mulDone
	VMOVDQU        (SI), X1
	VGF2P8AFFINEQB $0, X0, X1, X1
	VMOVDQU        X1, (DI)

mulDone:
	VZEROUPPER
	RET

// func gfniXorAVX2(matrix uint64, dst, src []byte)
TEXT ·gfniXorAVX2(SB), NOSPLIT, $0-56
	VPBROADCASTQ matrix+0(FP), Y0
	MOVQ         dst_base+8(FP), DI
	MOVQ         dst_len+16(FP), CX
	MOVQ         src_base+32(FP), SI
	CMPQ         CX, $32
	JB           xor16

xor32:
	VMOVDQU        (SI), Y1
	VGF2P8AFFINEQB $0, Y0, Y1, Y1
	VPXOR          (DI), Y1, Y1
	VMOVDQU        Y1, (DI)
	ADDQ           $32, SI
	ADDQ           $32, DI
	SUBQ           $32, CX
	CMPQ           CX, $32
	JAE            xor32

xor16:
	TESTQ          CX, CX
	JZ             xorDone
	VMOVDQU        (SI), X1
	VGF2P8AFFINEQB $0, X0, X1, X1
	VPXOR          (DI), X1, X1
	VMOVDQU        X1, (DI)

xorDone:
	VZEROUPPER
	RET

// func avx2Mul(tables *[32]byte, dst, src []byte)
TEXT ·avx2Mul(SB), NOSPLIT, $0-56
	MOVQ           tables+0(FP), AX
	MOVQ           dst_base+8(FP), DI
	MOVQ           dst_len+16(FP), CX
	MOVQ           src_base+32(FP), SI
	VBROADCASTI128 (AX), Y0
	VBROADCASTI128 16(AX), Y1
	MOVQ           $0x0f0f0f0f0f0f0f0f, BX
	MOVQ           BX, X2
	VPBROADCASTQ   X2, Y2
	CMPQ           CX, $32
	JB             mul16

mul32:
	VMOVDQU (SI), Y3
	VPSRLQ  $4, Y3, Y4
	VPAND   Y2, Y3, Y3
	VPAND   Y2, Y4, Y4
	VPSHUFB Y3, Y0, Y3
	VPSHUFB Y4, Y1, Y4
	VPXOR   Y4, Y3, Y3
	VMOVDQU Y3, (DI)
	ADDQ    $32, SI
	ADDQ    $32, DI
	SUBQ    $32, CX
	CMPQ    CX, $32
	JAE     mul32

mul16:
	TESTQ   CX, CX
	JZ      mulDone
	VMOVDQU (SI), X3
	VPSRLQ  $4, X3, X4
	VPAND   X2, X3, X3
	VPAND   X2, X4, X4
	VPSHUFB X3, X0, X3
	VPSHUFB X4, X1, X4
	VPXOR   X4, X3, X3
	VMOVDQU X3, (DI)

mulDone:
	VZEROUPPER
	RET

// func avx2Xor(tables *[32]byte, dst, src []byte)
TEXT ·avx2Xor(SB), NOSPLIT, $0-56
	MOVQ           tables+0(FP), AX
	MOVQ           dst_base+8(FP), DI
	MOVQ           dst_len+16(FP), CX
	MOVQ           src_base+32(FP), SI
	VBROADCASTI128 (AX), Y0
	VBROADCASTI128 16(AX), Y1
	MOVQ           $0x0f0f0f0f0f0f0f0f, BX
	MOVQ           BX, X2
	VPBROADCASTQ   X2, Y2
	CMPQ           CX, $32
	JB             xor16

xor32:
	VMOVDQU (SI), Y3
	VPSRLQ  $4, Y3, Y4
	VPAND   Y2, Y3, Y3
	VPAND   Y2, Y4, Y4
	VPSHUFB Y3, Y0, Y3
	VPSHUFB Y4, Y1, Y4
	VPXOR   Y4, Y3, Y3
	VPXOR   (DI), Y3, Y3
	VMOVDQU Y3, (DI)
	ADDQ    $32, SI
	ADDQ    $32, DI
	SUBQ    $32, CX
	CMPQ    CX, $32
	JAE     xor32

xor16:
	TESTQ   CX, CX
	JZ      xorDone
	VMOVDQU (SI), X3
	VPSRLQ  $4, X3, X4
	VPAND   X2, X3, X3
	VPAND   X2, X4, X4
	VPSHUFB X3, X0, X3
	VPSHUFB X4, X1, X4
	VPXOR   X4, X3, X3
	VPXOR   (DI), X3, X3
	VMOVDQU X3, (DI)

xorDone:
	VZEROUPPER
	RET
