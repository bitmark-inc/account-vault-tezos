package feralfilefeature

import "blockwatch.cc/tzgo/micheline"

func NewElt(l, r micheline.Prim) micheline.Prim {
	return micheline.Prim{Type: micheline.PrimBinary, OpCode: micheline.D_ELT, Args: []micheline.Prim{l, r}}
}
