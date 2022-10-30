package codegen

import . "luago/binchunk"

func toProto(fi *funcInfo) *Prototype {
	proto := &Prototype {
		LineDefined: uint32(fi.line),
		LastLineDefined: uint32(fi.lastLine),

		NumParams: byte(fi.numParams),
		MaxStackSize: byte(fi.maxRegs),
		Code: fi.insts,
		Constants: getConstants(fi),
		Upvalues: getUpvalues(fi),
		Protos: toProtos(fi.subFuncs),
		
		LineInfo: fi.lineNums, // for debug
		LocVars: getLocVars(fi), // for debug
		UpvalueNames: getUpvalueNames(fi), // for debug
	}

	if fi.line == 0 {
		proto.LastLineDefined = 0
	}

	if fi.isVararg {
		proto.IsVararg = 1
	}
	
	return proto
}

func toProtos(fis []*funcInfo) []*Prototype {
	protos := make([]*Prototype, len(fis))

	for i, fi := range fis {
		protos[i] = toProto(fi)
	}

	return protos
}

func getConstants(fi *funcInfo) []interface{} {
	consts := make([]interface{}, len(fi.constants))

	for k, idx := range fi.constants {
		consts[idx] = k
	}

	return consts
}

func getUpvalues(fi *funcInfo) []Upvalue {
	upvals := make([]Upvalue, len(fi.upvalues))
	
	for _, uv := range fi.upvalues {
		if uv.locVarSlot >= 0 { // instack
			upvals[uv.index] = Upvalue{1, byte(uv.locVarSlot),}
		} else {
			upvals[uv.index] = Upvalue{0, byte(uv.upvalIndex),}
		}
	}

	return upvals
}

func getLocVars(fi *funcInfo) []LocVar {
	locVars := make([]LocVar, len(fi.locVars))

	for i, locVar := range fi.locVars {
		locVars[i] = LocVar {
			VarName: locVar.name,
			StartPC: uint32(locVar.startPC),
			EndPC: uint32(locVar.endPC),
		}
	}
	return locVars
}

func getUpvalueNames(fi *funcInfo) []string {
	names := make([]string, len(fi.upvalues))

	for name, uv := range fi.upvalues {
		names[uv.index] = name
	}

	return names
}