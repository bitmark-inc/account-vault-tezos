package feralfilev1

import "blockwatch.cc/tzgo/micheline"

func newElt(l, r micheline.Prim) micheline.Prim {
	return micheline.Prim{Type: micheline.PrimBinary, OpCode: micheline.D_ELT, Args: []micheline.Prim{l, r}}
}
