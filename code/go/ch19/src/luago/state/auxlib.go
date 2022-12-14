package state

import (
	// "fmt"
	"io/ioutil"

	"luago/stdlib"
)

import . "luago/api"

func (self *luaState) TypeName2(idx int) string {
	return self.TypeName(self.Type(idx))
}

func (self *luaState) Len2(idx int) int64 {
	self.Len(idx)
	i, isNum := self.ToIntegerX(-1)
	if !isNum {
		self.Error2("object length is not an integer")
	}

	self.Pop(1)
	return i
}

func (self *luaState) CheckStack2(sz int, msg string) {
	if !self.CheckStack(sz) {
		if msg != "" {
			self.Error2("stack overflow (%s)", msg)
		} else {
			self.Error2("stack overflow")
		}
	}
}

func (self *luaState) Error2(fmt string, a ...interface{}) int {
	self.PushFString(fmt, a...)
	return self.Error()
}

func (self *luaState) ToString2(idx int) string {
	if self.CallMeta(idx, "__tostring") { /* metafiled? */
		if !self.IsString(-1) {
			self.Error2("'__tostring' must return a string")
		}
	} else {
		switch self.Type(idx) {
		case LUA_TNUMBER:
			if self.IsInteger(idx) {
				self.PushFString("%d", self.ToInteger(idx))
			} else {
				self.PushFString("%g", self.ToNumber(idx))
			}

		case LUA_TSTRING:
			self.PushValue(idx)

		case LUA_TBOOLEAN:
			if self.ToBoolean(idx) {
				self.PushString("true")
			} else {
				self.PushString("false")
			}

		case LUA_TNIL:
			self.PushString("nil")

		default:
			tt := self.GetMetafield(idx, "__name") /* try name */
			var kind string
			if LUA_TSTRING == tt {
				kind = self.CheckString(-1)
			} else {
				kind = self.TypeName2(idx)
			}

			self.PushFString("%s: %p", kind, self.ToPointer(idx))
			if tt != LUA_TNIL {
				self.Remove(-2) /* remove '__name' */
			}
		}
	}

	return self.CheckString(-1)
}

func (self *luaState) LoadString(s string) int {
	return self.Load([]byte(s), s, "bt")
}

func (self *luaState) LoadFileX(filename, mode string) int {
	if data, err := ioutil.ReadFile(filename); err == nil {
		return self.Load(data, "@" + filename, mode)
	}

	return LUA_ERRFILE
}

func (self *luaState) LoadFile(filename string) int {
	return self.LoadFileX(filename, "bt")
}

func (self *luaState) DoString(str string) bool {
	return self.LoadString(str) == LUA_OK &&
			self.PCall(0, LUA_MULTRET, 0) == LUA_OK
}

func (self *luaState) DoFile(filename string) bool {
	return self.LoadFile(filename) == LUA_OK &&
			self.PCall(0, LUA_MULTRET, 0) == LUA_OK
}

func (self *luaState) ArgError(arg int, extraMsg string) int {
	return self.Error2("bad argument #%d (%s)", arg, extraMsg)
}

func (self *luaState) ArgCheck(cond bool, arg int, extraMsg string) {
	if !cond {
		self.ArgError(arg, extraMsg)
	}
}

func (self *luaState) CheckAny(arg int) {
	if self.Type(arg) == LUA_TNONE {
		self.ArgError(arg, "value expected")
	}
}

func (self *luaState) CheckType(arg int, t LuaType) {
	if self.Type(arg) != t {
		self.tagError(arg, t)
	}
}

func (self *luaState) CheckInteger(arg int) int64 {
	i, ok := self.ToIntegerX(arg)
	if !ok {
		self.tagError(arg, LUA_TNUMBER)
	}

	return i
}

func (self *luaState) CheckNumber(arg int) float64 {
	f, ok := self.ToNumberX(arg)
	if !ok {
		self.tagError(arg, LUA_TNUMBER)
	}

	return f
}

func (self *luaState) CheckString(arg int) string {
	s, ok := self.ToStringX(arg)
	if !ok {
		self.tagError(arg, LUA_TSTRING)
	}

	return s
}

func (self *luaState) OptInteger(arg int, def int64) int64 {
	if self.IsNoneOrNil(arg) {
		return def
	}

	return self.CheckInteger(arg)
}

func (self *luaState) OptNumber(arg int, def float64) float64 {
	if self.IsNoneOrNil(arg) {
		return def
	}

	return self.CheckNumber(arg)
}

func (self *luaState) OptString(arg int, def string) string {
	if self.IsNoneOrNil(arg) {
		return def
	}

	return self.CheckString(arg)
}

func (self *luaState) tagError(arg int, tag LuaType) {
	self.typeError(arg, self.TypeName(tag))
}

func (self *luaState) typeError(arg int, tname string) int {
	var typeArg string /* name for the type fo the actual argument */

	if self.GetMetafield(arg, "__name") == LUA_TSTRING {
		typeArg = self.ToString(-1) /* use the given type name */
	} else if self.Type(arg) == LUA_TLIGHTUSERDATA {
		typeArg = "light userdata" /* special name for messages */
	} else {
		typeArg = self.TypeName2(arg) /* standard name */
	}

	msg := tname + " expected, got " + typeArg
	return self.ArgError(arg, msg)
}

func (self *luaState) OpenLibs() {
	libs := []struct {
		name string
		openf GoFunction
	} {
		{"_G", stdlib.OpenBaseLib,},
		{"math", stdlib.OpenMathLib,},
		{"table", stdlib.OpenTableLib,},
		{"string", stdlib.OpenStringLib,},
		{"utf8", stdlib.OpenUTF8Lib,},
		{"os", stdlib.OpenOSLib,},
	}

	for _, l := range libs {
		self.RequireF(l.name, l.openf, true)
		self.Pop(1)
	}
}

func (self *luaState) RequireF(modname string, openf GoFunction, glb bool) {
	self.GetSubTable(LUA_REGISTRYINDEX, "_LOADED")
	self.GetField(-1, modname) /* LOADED[modname] */
	if !self.ToBoolean(-1) { /* package not already loaded? */
		self.Pop(1) /* remove field */

		self.PushGoFunction(openf)
		self.PushString(modname) /* argument to open function */
		self.Call(1, 1) /* call 'openf' to open module */
		self.PushValue(-1) /* make copy of module (call result) */
		self.SetField(-3, modname) /* _LOADEd[modname] = module */
	}

	self.Remove(-2) /* remove _LOADED table */

	if glb {
		self.PushValue(-1) /* copy of module */
		self.SetGlobal(modname) /* _G[modname] = module */
	}
}

func (self *luaState) GetSubTable(idx int, fname string) bool {
	if self.GetField(idx, fname) == LUA_TTABLE {
		return true /* table already there */
	}

	self.Pop(1) /* remove previous result */
	idx = self.AbsIndex(idx)
	self.NewTable()
	self.PushValue(-1) /* copy to be left at top */
	self.SetField(idx, fname) /* assign new table to field */
	return false /* false, because did not find table there */
}

func (self *luaState) GetMetafield(idx int, e string) LuaType {
	if !self.GetMetatable(idx) { /* no metatable? */
		return LUA_TNIL
	}

	self.PushString(e)
	tt := self.RawGet(-2)
	if tt == LUA_TNIL {
		self.Pop(2) /* remove metatable and metafield */
	} else {
		self.Remove(-2) /* remove only metatable */
	}

	return tt /* return metafield type */
}

func (self *luaState) CallMeta(obj int, e string) bool {
	obj = self.AbsIndex(obj)

	if self.GetMetafield(obj, e) == LUA_TNIL { /* no metafield? */
		return false
	}

	self.PushValue(obj)
	self.Call(1, LUA_MULTRET)
	return true
}

func (self *luaState) NewLib(l []FuncReg) {
	self.NewLibTable(l)
	self.SetFuncs(l, 0)
}

func (self *luaState) NewLibTable(l []FuncReg) {
	self.CreateTable(0, len(l))
}

func (self *luaState) SetFuncs(l []FuncReg, nup int) {
	self.CheckStack2(nup, "too many upvalues")

	for _, l1 := range l { /* fill the table with given functions */
		for i := 0; i < nup; i++ { /* copy upvalues to the top */
			self.PushValue(-nup)
		}

		// r[-(nup + 2)][name] = func
		self.PushGoClosure(l1.Func, nup) /* closure with those upvalues */
		self.SetField(-(nup + 2), l1.Name)
	}

	self.Pop(nup) /* remove upvaluse */
}
