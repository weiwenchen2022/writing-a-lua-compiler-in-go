package codegen

import . "luago/compiler/ast"
import . "luago/binchunk"

func GenProto(chunk *Block) *Prototype {
	fd := &FuncDefExp{
		LastLine: chunk.LastLine,
		IsVararg: true,
		Block: chunk,
	}
	
	fi := newFuncInfo(nil, fd)
	fi.addLocVar("_ENV")
	cgFuncDefExp(fi, fd, 0)
	
	return toProto(fi.subFuncs[0])
}