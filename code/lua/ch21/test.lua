function permgen(a, n)
	n = n or #a -- default for 'n' is size of 'a'
	if n <= 1 then -- nothing to change?
		coroutine.yield(a)
	else
		for i = 1, n do
			-- put i-th element as the last one
			a[n], a[i] = a[i], a[n]

			-- generate all permutations of the other elements
			permgen(a, n - 1)

			-- restore i-th element
			a[n], a[i] = a[i], a[n]
		end
	end
end

function permutations(a)
	return coroutine.wrap(function() permgen(a) end)
	-- local co = coroutine.create(function() permgen(a) end)
	-- return function() -- iterator
	-- 	local _, res = coroutine.resume(co)
	-- 	return res
	-- end
end

function printResult(a)
	print(table.concat(a, ", "))
end

for p in permutations {"a", "b", "c",} do
	printResult(p)
end
-- co = coroutine.create(function() print "Hello" end)
-- print(type(co))

-- main = coroutine.running()
-- print(type(main))
-- print(coroutine.status(main))

-- co = coroutine.create(function()
-- 	print(coroutine.status(co))

-- 	coroutine.resume(coroutine.create(function()
-- 		print(coroutine.status(co))
-- 	end))
-- end)

-- print(coroutine.status(co))
-- coroutine.resume(co)
-- print(coroutine.status(co))

-- co = coroutine.create(function(...)
-- 	print(...)

-- 	while true do
-- 		print(coroutine.yield())
-- 	end
-- end)

-- coroutine.resume(co, 1, 2, 3)
-- coroutine.resume(co, 4, 5, 6)
-- coroutine.resume(co, 7, 8, 9)

-- co = coroutine.create(function()
-- 	for k, v in pairs {"a", "b", "c",} do
-- 		coroutine.yield(k, v)
-- 	end

-- 	return "d", 4
-- end)

-- print(coroutine.resume(co))
-- print(coroutine.resume(co))
-- print(coroutine.resume(co))
-- print(coroutine.resume(co))
-- print(coroutine.resume(co))