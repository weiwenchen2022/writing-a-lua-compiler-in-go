package codegen

import . "luago/compiler/lexer"
import . "luago/compiler/ast"
import . "luago/vm"

// kind of operands
const (
	ARG_CONST = 1 // const index
	ARG_REG = 2 // register index
	ARG_UPVAL = 4 // upval index

	ARG_RK = ARG_REG | ARG_CONST
	ARG_RU = ARG_REG | ARG_UPVAL
	ARG_RUK = ARG_REG | ARG_UPVAL | ARG_CONST
)

func cgExp(fi *funcInfo, node Exp, a, n int) {
	line := lineOf(node)
	switch exp := node.(type) {
	case *NilExp:
		fi.emitLoadNil(line, a, n)

	case *FalseExp:
		fi.emitLoadBool(line, a, 0, 0)

	case *TrueExp:
		fi.emitLoadBool(line, a, 1, 0)

	case *IntegerExp:
		fi.emitLoadK(line, a, exp.Val)

	case *FloatExp:
		fi.emitLoadK(line, a, exp.Val)

	case *StringExp:
		fi.emitLoadK(line, a, exp.Str)

	case *ParensExp:
		cgExp(fi, exp.Exp, a, 1)

	case *VarargExp:
		cgVarargExp(fi, exp, a, n)

	case *FuncDefExp:
		cgFuncDefExp(fi, exp, a)

	case *TableConstructorExp:
		cgTableConstructorExp(fi, exp, a)

	case *UnopExp:
		cgUnopExp(fi, exp, a)

	case *BinopExp:
		cgBinopExp(fi, exp, a)

	case *ConcatExp:
		cgConcatExp(fi, exp, a)

	case *NameExp:
		cgNameExp(fi, exp, a)

	case *TableAccessExp:
		cgTableAccessExp(fi, exp, a)

	case *FuncCallExp:
		cgFuncCallExp(fi, exp, a, n)
	}
}

func cgVarargExp(fi *funcInfo, node *VarargExp, a, n int) {
	if !fi.isVararg {
		panic("cannot use '...' outside a vararg function")
	}

	fi.emitVararg(lineOf(node), a, n)
}

// r[a] := function(args) block end
func cgFuncDefExp(fi *funcInfo, node *FuncDefExp, a int) {
	subFI := newFuncInfo(fi, node)
	fi.subFuncs = append(fi.subFuncs, subFI)

	for _, param := range node.ParList {
		subFI.addLocVar(param, 0)
	}

	cgBlock(subFI, node.Block)
	subFI.exitScope(subFI.pc() + 2)
	subFI.emitReturn(lastLineOf(node), 0, 0)

	bx := len(fi.subFuncs) - 1
	fi.emitClosure(lastLineOf(node), a, bx)
}

func cgTableConstructorExp(fi *funcInfo, node *TableConstructorExp, a int) {
	nArr := 0
	for _, keyExp := range node.KeyExps {
		if keyExp == nil {
			nArr++
		}
	}

	nExps := len(node.KeyExps)
	multRet := nExps > 0 && isVarargOrFuncCall(node.ValExps[nExps - 1])

	fi.emitNewTable(lineOf(node), a, nArr, nExps - nArr)

	arrIdx := 0
	for i, keyExp := range node.KeyExps {
		valExp := node.ValExps[i]

		if keyExp == nil {
			arrIdx++
			tmp := fi.allocReg()
			if nExps - 1 == i && multRet {
				cgExp(fi, valExp, tmp, -1)
			} else {
				cgExp(fi, valExp, tmp, 1)
			}

			if arrIdx % LFIELDS_PER_FLUSH == 0 || nArr == arrIdx {
				n := arrIdx % LFIELDS_PER_FLUSH
				if n == 0 {
					n = LFIELDS_PER_FLUSH
				}
				
				c := (arrIdx - 1) / LFIELDS_PER_FLUSH + 1
				fi.freeRegs(n)

				line := lastLineOf(valExp)
				if nExps - 1 == i && multRet {
					fi.emitSetList(line, a, 0, c)
				} else {
					fi.emitSetList(line, a, n, c)
				}
			}

			continue
		}

		b := fi.allocReg()
		cgExp(fi, keyExp, b, 1)

		c := fi.allocReg()
		cgExp(fi, valExp, c, 1)
		fi.freeRegs(2)

		line := lastLineOf(valExp)
		fi.emitSetTable(line, a, b, c)
	}
}

// r[a] := op exp
func cgUnopExp(fi *funcInfo, node *UnopExp, a int) {
	oldRegs := fi.usedRegs
	b, _ := expToOpArg(fi, node.Exp, ARG_REG)
	fi.emitUnaryOp(lineOf(node), node.Op, a, b)
	fi.usedRegs = oldRegs
}

// r[a] := exp1 op exp2
func cgBinopExp(fi *funcInfo, node *BinopExp, a int) {
	switch node.Op {
	case TOKEN_OP_AND, TOKEN_OP_OR:
		oldRegs := fi.usedRegs
		b, _ := expToOpArg(fi, node.Exp1, ARG_REG)
		fi.usedRegs = oldRegs

		line := lineOf(node)
		if TOKEN_OP_AND == node.Op {
			fi.emitTestSet(line, a, b, 0)
		} else {
			fi.emitTestSet(line, a, b, 1)
		}

		pcOfJmp := fi.emitJmp(line, 0, 0)

		b, _ = expToOpArg(fi, node.Exp2, ARG_REG)
		fi.usedRegs = oldRegs
		fi.emitMove(line, a, b)
		fi.fixSbx(pcOfJmp, fi.pc() - pcOfJmp)

	default:
		oldRegs := fi.usedRegs
		b, _ := expToOpArg(fi, node.Exp1, ARG_RK)
		c, _ := expToOpArg(fi, node.Exp2, ARG_RK)
		fi.emitBinaryOp(lineOf(node), node.Op, a, b, c)
		fi.usedRegs = oldRegs
	}
}

// r[a] := exp1 .. exp2
func cgConcatExp(fi *funcInfo, node *ConcatExp, a int) {
	for _, subExp := range node.Exps {
		a := fi.allocReg()
		cgExp(fi, subExp, a, 1)
	}

	c := fi.usedRegs - 1
	b := c - len(node.Exps) + 1
	fi.freeRegs(c - b + 1)
	fi.emitABC(lineOf(node), OP_CONCAT, a, b, c)
}

// r[a] := name
func cgNameExp(fi *funcInfo, node *NameExp, a int) {
	line := lineOf(node)

	if r := fi.slotOfLocVar(node.Name); r >= 0 {
		fi.emitMove(line, a, r)
	} else if idx := fi.indexOfUpval(node.Name); idx >= 0 {
		fi.emitGetUpval(line, a, idx)
	} else { // x => _ENV['x']
		taExp := &TableAccessExp {
			LastLine: line,
			PrefixExp: &NameExp{0,  "_ENV",},
			KeyExp: &StringExp{0, node.Name,},
		}

		cgTableAccessExp(fi, taExp, a)
	}
}

// r[a] := prefix[key]
func cgTableAccessExp(fi *funcInfo, node *TableAccessExp, a int) {
	oldRegs := fi.usedRegs
	b, kindB := expToOpArg(fi, node.PrefixExp, ARG_RU)
	c, _ := expToOpArg(fi, node.KeyExp, ARG_RK)
	fi.usedRegs = oldRegs

	line := lastLineOf(node)
	if ARG_UPVAL == kindB {
		fi.emitGetTabUp(line, a, b, c)
	} else {
		fi.emitGetTable(line, a, b, c)
	}
}

// r[a] := f(args)
func cgFuncCallExp(fi *funcInfo, node *FuncCallExp, a, n int) {
	nArgs := prepFuncCall(fi, node, a)
	fi.emitCall(lineOf(node), a, nArgs, n)
}

// return f(args)
func cgTailCallExp(fi *funcInfo, node *FuncCallExp, a int) {
	nArgs := prepFuncCall(fi, node, a)
	fi.emitTailCall(lineOf(node), a, nArgs)
}

func prepFuncCall(fi *funcInfo, node *FuncCallExp, a int) int {
	nArgs := len(node.Args)
	lastArgIsVarargOrFuncCall := false

	cgExp(fi, node.PrefixExp, a, 1)
	if node.NameExp != nil {
		fi.allocReg()
		c, k := expToOpArg(fi, node.NameExp, ARG_RK)
		fi.emitSelf(lineOf(node), a, a, c)
		if ARG_REG == k {
			fi.freeReg()
		}
	}

	for i, arg := range node.Args {
		tmp := fi.allocReg()
		if nArgs - 1 == i && isVarargOrFuncCall(arg) {
			lastArgIsVarargOrFuncCall = true
			cgExp(fi, arg, tmp, -1)
		} else {
			cgExp(fi, arg, tmp, 1)
		}
	}
	fi.freeRegs(nArgs)

	if node.NameExp != nil {
		fi.freeReg()
		nArgs++
	}
	
	if lastArgIsVarargOrFuncCall {
		nArgs = - 1
	}

	return nArgs
}

func expToOpArg(fi *funcInfo, node Exp, argKinds int) (arg, argKind int) {
	if argKinds & ARG_CONST != 0 {
		idx := -1

		switch x := node.(type) {
		case *NilExp:
			idx = fi.indexOfConstant(nil)

		case *FalseExp:
			idx = fi.indexOfConstant(false)

		case *TrueExp:
			idx = fi.indexOfConstant(true)

		case *IntegerExp:
			idx = fi.indexOfConstant(x.Val)

		case *FloatExp:
			idx = fi.indexOfConstant(x.Val)

		case *StringExp:
			idx = fi.indexOfConstant(x.Str)
		}

		if 0 <= idx && idx <= 0xFF {
			return 0x100 + idx, ARG_CONST
		}
	}

	if nameExp, ok := node.(* NameExp); ok {
		if argKinds & ARG_REG != 0 {
			if r := fi.slotOfLocVar(nameExp.Name); r >= 0 {
				return r, ARG_REG
			}
		}

		if argKinds & ARG_UPVAL != 0 {
			if idx := fi.indexOfUpval(nameExp.Name); idx >= 0 {
				return idx, ARG_UPVAL
			}
		}
	}

	a := fi.allocReg()
	cgExp(fi, node, a, 1)
	return a, ARG_REG
}