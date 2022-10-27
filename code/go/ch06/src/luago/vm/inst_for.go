package vm

import . "luago/api"

// R(A) -= R(A + 2); pc += sBx
func forPrep(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	// R(A) -= R(A+2)
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(LUA_OPSUB)
	vm.Replace(a)

	// pc += sBx
	vm.AddPC(sBx)
}

// R(A) += R(A + 2)
// if R(A) <?= $(A + 1) then {
//   pc += sBx; R(A + 3) = R(A)
// }
func forLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	// R(A) += R(A+2)
	vm.PushValue(a + 2)
	vm.PushValue(a)
	vm.Arith(LUA_OPADD)
	vm.Replace(a)

	// R(A) <?= R(A+1)
	isPositiveStep := vm.ToNumber(a + 2) >= 0
	if isPositiveStep && vm.Compare(a, a + 1, LUA_OPLE) ||
		!isPositiveStep && vm.Compare(a + 1, a, LUA_OPLE) {
			// pc += sBx; R(A + 3) = R(A)
			vm.AddPC(sBx)
			vm.Copy(a, a + 3)
		}
}