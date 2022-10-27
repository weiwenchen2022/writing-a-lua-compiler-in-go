local function max(...)
	local args = {...}
	local val, idx
	for i = 1, #args do
		if val == nil or val < args[i] then
			val, idx = args[i], i
		end
	end

	return val, idx
end

local function assert(v)
	if not v then fail() end
end

local v1 = max(3, 9, 7, 128, 35)
assert(128 == v1)

local v2, i2 = max(3, 9, 7, 128, 35)
assert(128 == v2 and 4 == i2)

local v3, i3 = max(max(3, 9, 7, 128, 35))
assert(128 == v3 and 1 == i3)

local t = {max(3, 9, 7, 128, 35),}
assert(128 == t[1] and 4 == t[2])