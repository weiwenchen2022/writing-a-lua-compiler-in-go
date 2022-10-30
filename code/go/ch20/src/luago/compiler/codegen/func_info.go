package codegen

import . "luago/compiler/lexer"
import . "luago/compiler/ast"
import . "luago/vm"

var arithAndBitwiseBinops = map[int]int {
	TOKEN_OP_ADD: OP_ADD,
	TOKEN_OP_SUB: OP_SUB,
	TOKEN_OP_MUL: OP_MUL,
	TOKEN_OP_MOD: OP_MOD,
	TOKEN_OP_POW: OP_POW,
	TOKEN_OP_DIV: OP_DIV,
	TOKEN_OP_IDIV: OP_IDIV,

	TOKEN_OP_BAND: OP_BAND,
	TOKEN_OP_BOR: OP_BOR,
	TOKEN_OP_BXOR: OP_BXOR,
	TOKEN_OP_SHL: OP_SHL,
	TOKEN_OP_SHR: OP_SHR,
}

type locVarInfo struct {
	prev *locVarInfo
	name string
	scopeLv int
	slot int
	captured bool

	startPC int
	endPC int
}

type upvalInfo struct {
	locVarSlot int
	upvalIndex int
	index int
}

type funcInfo struct {
	parent *funcInfo
	subFuncs []*funcInfo

	usedRegs int
	maxRegs int

	scopeLv int
	locVars []*locVarInfo
	locNames map[string]*locVarInfo

	upvalues map[string]upvalInfo

	constants map[interface{}]int
	
	breaks [][]int

	insts []uint32

	lineNums []uint32
	line int
	lastLine int

	numParams int
	isVararg bool
}

func newFuncInfo(parent *funcInfo, fd *FuncDefExp) *funcInfo {
	return &funcInfo {
		parent: parent,
		subFuncs: []*funcInfo{},

		locVars: make([]*locVarInfo, 0, 8),
		locNames: map[string]*locVarInfo{},

		upvalues: map[string]upvalInfo{},

		constants: map[interface{}]int{},
		
		breaks: make([][]int, 1),
		
		insts: make([]uint32, 0, 8),

		lineNums: make([]uint32, 0, 8),
		line: lineOf(fd),
		lastLine: lastLineOf(fd),

		numParams: len(fd.ParList),
		isVararg: fd.IsVararg,
	}
}

/* constants */
func (self *funcInfo) indexOfConstant(k interface{}) int {
	if idx, found := self.constants[k]; found {
		return idx
	}

	idx := len(self.constants)
	self.constants[k] = idx
	return idx
}

/* registers */
func (self *funcInfo) allocReg() int {
	self.usedRegs++
	if self.usedRegs >= 255 {
		panic("function or expression needs too many register")
	}

	if self.maxRegs < self.usedRegs {
		self.maxRegs = self.usedRegs
	}

	return self.usedRegs - 1
}

func (self *funcInfo) freeReg() {
	if self.usedRegs <= 0 {
		panic("usedRegs <= 0!")
	}

	self.usedRegs--
}

func (self *funcInfo) allocRegs(n int) int {
	if n <= 0 {
		panic("n <= 0!")
	}

	for i := 0; i < n; i++ {
		self.allocReg()
	}

	return self.usedRegs - n
}

func (self *funcInfo) freeRegs(n int) {
	if n < 0 {
		panic("n < 0!")
	}

	for i := 0; i < n; i++ {
		self.freeReg()
	}
}

/* lexical scope */
func (self *funcInfo) enterScope(breakable bool) {
	self.scopeLv++

	if breakable {
		self.breaks = append(self.breaks, []int{})
	} else {
		self.breaks = append(self.breaks, nil)
	}
}

func (self *funcInfo) exitScope(endPC int) {
	pendingBreakJmps := self.breaks[len(self.breaks) - 1]
	self.breaks = self.breaks[:len(self.breaks) - 1]

	// a := self.getJmpArgA()
	for _, pc := range pendingBreakJmps {
		sBx := self.pc() - pc
		self.fixSbx(pc, sBx)
		// i := (sBx + MAXARG_sBx) << 14 | a << 6 | OP_JMP
		// self.insts[pc] = uint32(i)
	}

	self.scopeLv--
	for _, locVar := range self.locNames {
		if self.scopeLv < locVar.scopeLv { // out of scope
			locVar.endPC = endPC
			self.removeLocVar(locVar)
		}
	}
}

func (self *funcInfo) removeLocVar(locVar *locVarInfo) {
	self.freeReg()

	if locVar.prev == nil {
		delete(self.locNames, locVar.name)
	} else if locVar.scopeLv == locVar.prev.scopeLv {
		self.removeLocVar(locVar.prev)
	} else {
		self.locNames[locVar.name] = locVar.prev
	}
}

func (self *funcInfo) addLocVar(name string, startPC int) int {
	newVar := &locVarInfo {
		name: name,
		prev: self.locNames[name],
		scopeLv: self.scopeLv,
		slot: self.allocReg(),

		startPC: startPC,
		endPC: 0,
	}

	self.locVars = append(self.locVars, newVar)
	self.locNames[name] = newVar
	return newVar.slot
}

func (self *funcInfo) slotOfLocVar(name string) int {
	if locVar, found := self.locNames[name]; found {
		return locVar.slot
	}

	return -1
}

func (self *funcInfo) addBreakJmp(pc int) {
	for i := self.scopeLv; i >= 0; i-- {
		if self.breaks[i] != nil { // breakable
			self.breaks[i] = append(self.breaks[i], pc)
			return
		}
	}

	panic("<break> at line ? not inside a loop!")
}

/* upvalues */
func (self *funcInfo) indexOfUpval(name string) int {
	if upval, ok := self.upvalues[name]; ok {
		return upval.index
	}

	if self.parent != nil {
		if locVar, found := self.parent.locNames[name]; found {
			idx := len(self.upvalues)
			self.upvalues[name] = upvalInfo{locVar.slot, -1, idx,}
			locVar.captured = true

			return idx
		}

		if uvIdx := self.parent.indexOfUpval(name); uvIdx >= 0 {
			idx := len(self.upvalues)
			self.upvalues[name] = upvalInfo{-1, uvIdx, idx,}

			return idx
		}
	}

	return -1
}

func (self *funcInfo) closeOpenUpvals(line int) {
	a := self.getJmpArgA()
	if a > 0 {
		self.emitJmp(line, a, 0)
	}
}

func (self *funcInfo) getJmpArgA() int {
	hasCaptureLocVars := false
	minSlotOfLocVars := self.maxRegs

	for _, locVar := range self.locNames {
		if self.scopeLv == locVar.scopeLv {
			for v := locVar; v != nil && self.scopeLv == v.scopeLv; v = v.prev {
				if v.captured {
					hasCaptureLocVars = true
				}

				if v.slot < minSlotOfLocVars && v.name[0] != '(' {
					minSlotOfLocVars = v.slot
				}
			}	
		}
	}

	if hasCaptureLocVars {
		return minSlotOfLocVars + 1
	} else {
		return 0
	}
}

/* code */
func (self *funcInfo) pc() int {
	return len(self.insts) - 1
}

func (self *funcInfo) fixSbx(pc, sBx int) {
	i := self.insts[pc]
	i = uint32(sBx + MAXARG_sBx) << 14 | (i & 0x3FFF)
	self.insts[pc] = i
}

func (self *funcInfo) fixEndPC(name string, delta int) {
	if locVar := self.locNames[name]; locVar != nil {
		locVar.endPC += delta
	}
}

func (self *funcInfo) emitABC(line, opcode, a, b, c int) int {
	i := b << 23 | c << 14 | a << 6 | opcode
	self.insts = append(self.insts, uint32(i))
	self.lineNums = append(self.lineNums, uint32(line))

	return self.pc()
}

func (self *funcInfo) emitABx(line, opcode, a, bx int) int {
	i := bx << 14 | a << 6 | opcode
	self.insts = append(self.insts, uint32(i))
	self.lineNums = append(self.lineNums, uint32(line))

	return self.pc()
}

func (self *funcInfo) emitAsBx(line, opcode, a, b int) int {
	return self.emitABx(line, opcode, a, b + MAXARG_sBx)
}

func (self *funcInfo) emitAx(line, opcode, ax int) int {
	i := ax << 6 | opcode
	self.insts = append(self.insts, uint32(i))
	self.lineNums = append(self.lineNums, uint32(line))

	return self.pc()
}

// r[a] := r[b]
func (self *funcInfo) emitMove(line, a, b int) {
	self.emitABC(line, OP_MOVE, a, b, 0)
}

// r[a], r[a+1], ..., r[a+b] := nil
func (self *funcInfo) emitLoadNil(line, a, n int) {
	self.emitABC(line, OP_LOADNIL, a, n-1, 0)
}

// r[a] := (bool)b; if (c) pc++
func (self *funcInfo) emitLoadBool(line, a, b, c int) {
	self.emitABC(line, OP_LOADBOOL, a, b, c)
}

// r[a] := kst[bx]
func (self *funcInfo) emitLoadK(line, a int, k interface{}) {
	idx := self.indexOfConstant(k)
	if idx < (1 << 18) {
		self.emitABx(line, OP_LOADK, a, idx)
	} else {
		self.emitABx(line, OP_LOADKX, a, 0)
		self.emitAx(line, OP_EXTRAARG, idx)
	}
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (self *funcInfo) emitVararg(line, a, n int) {
	self.emitABC(line, OP_VARARG, a, n+1, 0)
}

// r(a) := closure(KPROTO[bx])
func (self *funcInfo) emitClosure(line, a, bx int) {
	self.emitABx(line, OP_CLOSURE, a, bx)
}

// r(A) := {} (size = b,c)
func (self *funcInfo) emitNewTable(line, a, nArr, nRec int) {
	self.emitABC(line, OP_NEWTABLE, a, Int2fb(nArr), Int2fb(nRec))
}

// r(a)[(c-1) * FPF + i] := r(a + i), 1 <= i <= b
func (self *funcInfo) emitSetList(line, a, b, c int) {
	self.emitABC(line, OP_SETLIST, a, b, c)
}

// r(a) := r(b)[rk(c)]
func (self *funcInfo) emitGetTable(line, a, b, c int) {
	self.emitABC(line, OP_GETTABLE, a, b, c)
}

// r(a)[rk(b)] := rk(c)
func (self *funcInfo) emitSetTable(line, a, b, c int) {
	self.emitABC(line, OP_SETTABLE, a, b, c)
}

// r(a) := upval[b]
func (self *funcInfo) emitGetUpval(line, a, b int) {
	self.emitABC(line, OP_GETUPVAL, a, b, 0)
}

// upval[b] := r(a)
func (self *funcInfo) emitSetUpval(line, a, b int) {
	self.emitABC(line, OP_SETUPVAL, a, b, 0)
}

// r(a) := upval[b][rk(c)]
func (self *funcInfo) emitGetTabUp(line, a, b, c int) {
	self.emitABC(line, OP_GETTABUP, a, b, c)
}

// upval[a][rk(b)] := rk(c)
func (self *funcInfo) emitSetTabUp(line, a, b, c int) {
	self.emitABC(line, OP_SETTABUP, a, b, c)
}

// r(a), ..., r(a+c-2) := r(a)(r(a+1), ..., R(a+b-1))
func (self *funcInfo) emitCall(line, a, nArgs, nRet int) {
	self.emitABC(line, OP_CALL, a, nArgs + 1, nRet + 1)
}

// return r(a)(r(a+1), ..., r(a+b-1))
func (self *funcInfo) emitTailCall(line, a, nArgs int) {
	self.emitABC(line, OP_TAILCALL, a, nArgs + 1, 0)
}

// return r[a], ..., r(a + b - 2)
func (self *funcInfo) emitReturn(line, a, n int) {
	self.emitABC(line, OP_RETURN, a, n+1, 0)
}

// r(a+1) := r(b); r(a) := r(b)[rk(c)]
func (self *funcInfo) emitSelf(line, a, b, c int) {
	self.emitABC(line, OP_SELF, a, b, c)
}

// pc += sBx; if (a) close all upvalus >= r[a - 1]
func (self *funcInfo) emitJmp(line, a, sBx int) int {
	return self.emitAsBx(line, OP_JMP, a, sBx)
}

// if not (r(a) <=> c) then pc++
func (self *funcInfo) emitTest(line, a, c int) {
	self.emitABC(line, OP_TEST, a, 0, c)
}

// if (r(b) <=> c) then r(a) = r(b) else pc++
func (self *funcInfo) emitTestSet(line, a, b, c int) {
	self.emitABC(line, OP_TESTSET, a, b, c)
}

// r(a) -= r(a + 2); pc += sBx
func (self *funcInfo) emitForPrep(line, a, sBx int) int {
	return self.emitAsBx(line, OP_FORPREP, a, sBx)
}

// r(a) += r(a + 2); if r(a) <?= r(a+1) then { pc += sBx; r(a + 3) = r(a) }
func (self *funcInfo) emitForLoop(line, a, sBx int) int {
	return self.emitAsBx(line, OP_FORLOOP, a, sBx)
}

// r(a+3), ..., r(a+2+c) := r(a)(r(a+1), r(a+2))
func (self *funcInfo) emitTForCall(line, a, c int) int {
	return self.emitABC(line, OP_TFORCALL, a, 0, c)
}

// if r(a+1) ~= nil then { r(a) = r(a+1); pc += sBx }
func (self *funcInfo) emitTForLoop(line, a, sBx int) int {
	return self.emitAsBx(line, OP_TFORLOOP, a, sBx)
}

// r[a] := op r[b]
func (self *funcInfo) emitUnaryOp(line, op, a, b int) {
	switch op {
	case TOKEN_OP_NOT:
		self.emitABC(line, OP_NOT, a, b, 0)

	case TOKEN_OP_BNOT:
		self.emitABC(line, OP_BNOT, a, b, 0)

	case TOKEN_OP_LEN:
		self.emitABC(line, OP_LEN, a, b, 0)

	case TOKEN_OP_UNM:
		self.emitABC(line, OP_UNM, a, b, 0)
	}
}

// r[a] := rk[b] op rk[c]
// arith & bitwise & relational
func (self *funcInfo) emitBinaryOp(line, op, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		self.emitABC(line, opcode, a, b, c)
	} else {
		switch op {
		case TOKEN_OP_EQ:
			self.emitABC(line, OP_EQ, 1, b, c)

		case TOKEN_OP_NE:
			self.emitABC(line, OP_EQ, 0, b, c)

		case TOKEN_OP_LT:
			self.emitABC(line, OP_LT, 1, b, c)

		case TOKEN_OP_GT:
			self.emitABC(line, OP_LT, 1, c, b)

		case TOKEN_OP_LE:
			self.emitABC(line, OP_LE, 1, b, c)

		case TOKEN_OP_GE:
			self.emitABC(line, OP_LE, 1, c, b)

		}

		self.emitJmp(line, 0, 1)
		self.emitLoadBool(line, a, 0, 1)
		self.emitLoadBool(line, a, 1, 0)
	}
}