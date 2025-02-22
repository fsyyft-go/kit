//go:build arm64

#include "textflag.h"

// func getg() *g
TEXT ·getg(SB),NOSPLIT,$0-8
#ifdef GOARCH_386
	MOVL (TLS), AX
	MOVL AX, ret+0(FP)
#endif
#ifdef GOARCH_amd64
	MOVQ (TLS), AX
	MOVQ AX, ret+0(FP)
#endif
#ifdef GOARCH_arm
	MOVW g, ret+0(FP)
#endif
#ifdef GOARCH_arm64
	MOVD g, ret+0(FP)
#endif
#ifdef GOARCH_s390x
	MOVD g, ret+0(FP)
#endif
	RET
