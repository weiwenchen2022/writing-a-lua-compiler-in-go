package state

import . "luago/api"

type luaState struct {
	stack *luaStack
}

func New() LuaState {
	return &luaState {
		stack: newLuaStack(20),
	}
}