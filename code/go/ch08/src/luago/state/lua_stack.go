package state

type luaStack struct {
	slots []luaValue
	top int

	prev *luaStack
	closure *closure
	varargs []luaValue
	pc int
}

func newLuaStack(size int) *luaStack {
	return &luaStack {
		slots : make([]luaValue, size),
		top : 0,
	}
}

func (self *luaStack) check(n int) {
	free := len(self.slots) - self.top
	for i := free; i < n; i++ {
		self.slots = append(self.slots, nil)
	}
}

func (self *luaStack) push(val luaValue) {
	if len(self.slots) == self.top {
		panic("stack overflow!")
	}

	self.slots[self.top] = val
	self.top++
}

func (self *luaStack) pop() luaValue {
	if self.top < 1 {
		panic("stack underflow")
	}

	self.top--
	val := self.slots[self.top]
	self.slots[self.top] = nil
	return val
}

func (self *luaStack) absIndex(idx int) int {
	if idx >= 0 {
		return idx
	}

	return idx + self.top + 1
}

func (self *luaStack) isValid(idx int) bool {
	idx = self.absIndex(idx)
	return 0 < idx && idx <= self.top
}

func (self *luaStack) get(idx int) luaValue {
	idx = self.absIndex(idx)
	if 0 < idx && idx <= self.top {
		return self.slots[idx - 1]
	}

	return nil
}

func (self *luaStack) set(idx int, val luaValue) {
	idx = self.absIndex(idx)
	if 0 < idx && idx <= self.top {
		self.slots[idx - 1] = val
		return
	}

	panic("invalid index!")
}

func (self *luaStack) reverse(from, to int) {
	slots := self.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}

func (self *luaStack) popN(n int) []luaValue {
	vals := make([]luaValue, n)
	for i := n - 1; i >= 0; i-- {
		vals[i] = self.pop()
	}

	return vals
}

func (self *luaStack) pushN(vals []luaValue, n int) {
	nVals := len(vals)
	if n < 0 {
		n = nVals
	}
	
	for i := 0; i < n; i++ {
		if i < nVals {
			self.push(vals[i])
		} else {
			self.push(nil)
		}
	}
}