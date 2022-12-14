package state

import (
	"fmt"
)

import . "luago/api"

func (self *luaState) TypeName(tp LuaType) string {
	switch tp {
	case LUA_TNONE: return "no value"
	case LUA_TNIL: return "nil"
	case LUA_TBOOLEAN: return "boolean"
	case LUA_TNUMBER: return "number"
	case LUA_TSTRING: return "string"
	case LUA_TTABLE: return "table"
	case LUA_TFUNCTION: return "function"
	case LUA_TTHREAD: return "thread"
	default: return "userdata"
	}
}

func (self *luaState) Type(idx int) LuaType {
	if self.stack.isValid(idx) {
		val := self.stack.get(idx)
		return typeOf(val)
	}

	return LUA_TNONE
}

func (self *luaState) IsNone(idx int) bool {
	return LUA_TNONE == self.Type(idx)
}

func (self *luaState) IsNil(idx int) bool {
	return LUA_TNIL == self.Type(idx)
}

func (self *luaState) IsNoneOrNil(idx int) bool {
	return self.Type(idx) <= LUA_TNIL
}

func (self *luaState) IsBoolean(idx int) bool {
	return LUA_TBOOLEAN == self.Type(idx)
}

func (self *luaState) IsInteger(idx int) bool {
	val := self.stack.get(idx)
	_, ok := val.(int64)
	return ok;
}

func (self *luaState) IsNumber(idx int) bool {
	_, ok := self.ToNumberX(idx)
	return ok
}

func (self *luaState) IsString(idx int) bool {
	t := self.Type(idx)
	return LUA_TSTRING == t || LUA_TNUMBER == t
}

func (self *luaState) IsGoFunction(idx int) bool {
	val := self.stack.get(idx)
	if c, ok := val.(* closure); ok {
		return c.goFunc != nil
	}

	return false
}

func (self *luaState) IsFunction(idx int) bool {
	return LUA_TFUNCTION == self.Type(idx)
}

func (self *luaState) ToBoolean(idx int) bool {
	val := self.stack.get(idx)
	return convertToBoolean(val)
}

func (self *luaState) ToInteger(idx int) int64 {
	i, _ := self.ToIntegerX(idx)
	return i
}

func (self *luaState) ToIntegerX(idx int) (int64, bool) {
	val := self.stack.get(idx)
	return convertToInteger(val)
}

func (self *luaState) ToNumber(idx int) float64 {
	n, _ := self.ToNumberX(idx)
	return n
}

func (self *luaState) ToNumberX(idx int) (float64, bool) {
	val := self.stack.get(idx)
	return convertToFloat(val)
}

func (self *luaState) ToString(idx int) string {
	s, _ := self.ToStringX(idx)
	return s
}

func (self *luaState) ToStringX(idx int) (string, bool) {
	val := self.stack.get(idx)
	switch x := val.(type) {
	case string: return x, true
	case int64, float64:
		s := fmt.Sprintf("%v", x)
		self.stack.set(idx, s)
		return s, true
	default: return "", false
	}
}

func (self *luaState) ToGoFunction(idx int) GoFunction {
	val := self.stack.get(idx)
	if c, ok := val.(* closure); ok {
		return c.goFunc
	}

	return nil
}

// [-0, +0, ???]
// http://www.lua.org/manual/5.3/manual.html#lua_tothread
func (self *luaState) ToThread(idx int) LuaState {
	if val := self.stack.get(idx); val != nil {
		if ls, ok := val.(* luaState); ok {
			return ls
		}
	}

	return nil
}

// [-0, +0, ???]
// http://www.lua.org/manual/5.3/manual.html#lua_topointer
func (self *luaState) ToPointer(idx int) interface{} {
	// todo
	return self.stack.get(idx)
}

func (self *luaState) RawLen(idx int) uint {
	val := self.stack.get(idx)
	switch x := val.(type) {
	case string:
		return uint(len(x))

	case *luaTable:
		return uint(x.len())

	default:
		return 0

	}
}