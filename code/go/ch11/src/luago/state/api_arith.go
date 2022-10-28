package state

import (
	"math"

	"luago/number"
)

import . "luago/api"

var (
	iadd = func(a, b int64) int64 { return a + b }
	fadd = func(a, b float64) float64 { return a + b }

	isub = func(a, b int64) int64 { return a - b }
	fsub = func(a, b float64) float64 { return a - b }

	imul = func(a, b int64) int64 { return a * b }
	fmul = func(a, b float64) float64 { return a * b }

	imod = number.IMod
	fmod = number.FMod

	pow = math.Pow
	div = func(a, b float64) float64 { return a / b }

	iidiv = number.IFloorDiv
	fidiv = number.FFloorDiv

	band = func(a, b int64) int64 { return a & b }
	bor = func(a, b int64) int64 { return a | b }
	bxor = func(a, b int64) int64 { return a ^ b }

	shl = number.ShiftLeft
	shr = number.ShiftRight

	iunm = func(a, _ int64) int64 { return -a }
	funm = func(a, _ float64) float64 { return -a }

	bnot = func(a, _ int64) int64 { return ^a }
)

type operator struct {
	integerFunc func(int64, int64) int64
	floatFunc func(float64, float64) float64

	metamethod string
}

var operators = []operator {
	operator{iadd, fadd, "__add"},
	operator{isub, fsub, "__sub"},
	operator{imul, fmul, "__mul"},
	operator{imod, fmod, "__mod"},
	operator{nil, pow, "__pow"},
	operator{nil, div, "__div"},
	operator{iidiv, fidiv, "__idiv"},

	operator{band, nil, "__band"},
	operator{bor, nil, "__bor"},
	operator{bxor, nil, "__bxor"},
	
	operator{shl, nil, "__shl"},
	operator{shr, nil, "__shr"},

	operator{iunm, funm, "__unm"},
	operator{bnot, nil, "__bnot"},
}

// [-(2|1), +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_arith
func (self *luaState) Arith(op ArithOp) {
	var a, b luaValue // operands
	b = self.stack.pop()
	if op != LUA_OPUNM && op != LUA_OPBNOT {
		a = self.stack.pop()
	} else {
		a = b
	}

	operator := operators[op]
	if result := _arith(a, b, operator); result != nil {
		self.stack.push(result)
		return
	}

	mm := operator.metamethod
	if result, ok := callMetamethod(a, b, mm, self); ok {
		self.stack.push(result)
		return
	}
	
	panic("arithmetic error!")
}

func _arith(a, b luaValue, op operator) luaValue {
	if op.floatFunc == nil { // bitwise
		if x, ok := convertToInteger(a); ok {
			if y, ok := convertToInteger(b); ok {
				return op.integerFunc(x, y)
			}
		}
	} else { // arith
		if op.integerFunc != nil { // add, sub, mul, mod, idiv, unm
			if x, ok := a.(int64); ok {
				if y, ok := b.(int64); ok {
					return op.integerFunc(x, y)
				}
			}
		}

		if x, ok := convertToFloat(a); ok {
			if y, ok := convertToFloat(b); ok {
				return op.floatFunc(x, y)
			}
		}
	}

	return nil
}










