package state

import . "luago/api"

type luaState struct {
	stack *luaStack

	registry *luaTable

}

func New() *luaState {
	registry := newLuaTable(2, 0)
	registry.put(LUA_RIDX_GLOBALS, newLuaTable(0, 0))

	ls := &luaState {registry: registry,}
	ls.pushLuaStack(newLuaStack(LUA_MINSTACK, ls))

	return ls
}

func (self *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = self.stack
	self.stack = stack
}

func (self *luaState) popLuaStack() {
	stack := self.stack
	self.stack = stack.prev
	stack.prev = nil
}