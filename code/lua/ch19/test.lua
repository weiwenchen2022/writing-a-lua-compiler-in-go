
-- test math lib
print(math.type(100))
print(math.type(3.14))
print(math.type("100"))
print(math.tointeger(100.0))
print(math.tointeger("100.0"))
print(math.tointeger(3.14))

-- test table lib
t = table.pack(1, 2, 3, 4, 5); print(table.unpack(t))
table.move(t, 4, 5, 1); print(table.unpack(t))
table.insert(t, 3, 2); print(table.unpack(t))
table.remove(t, 2); print(table.unpack(t))
table.sort(t); print(table.unpack(t))
print(table.concat(t, ", "))

-- test string lib
print(string.len "abc")
print(string.rep("a", 3, ", "))
print(string.reverse "abc")
print(string.lower "ABC")
print(string.upper "abc")
print(string.sub("abcdefg", 3, 5))
print(string.byte("abcdefg", 3, 5))
print(string.char(99, 100, 101))

s = "aBc"
print(s:len())
print(s:rep(3, ", "))
print(s:reverse())
print(s:upper())
print(s:lower())
print(s:sub(1, 2))
print(s:byte(1, 2))

-- test utf8 lib
print(utf8.char(0x4f60, 0x597d))
print(utf8.len "你好，世界！")

-- test OS lib
print(os.time())
print(os.time {year = 2018, month = 2, day = 14,
hour = 12, min = 30, sec = 30})

print(os.date())
t = os.date("*t")
print(t.year)
print(t.month)
print(t.day)
print(t.hour)
print(t.min)
print(t.sec)

local x = os.clock()
local s = 0;
for i = 1, 1000000 do s = s + i end
print(string.format("elapsed time: %.2f\n", os.clock() - x))

local t5_3 = os.time {year = 2015, month = 1, day = 12,}
local t5_2 = os.time {year = 2011, month = 12, day = 16,}
local d = os.difftime(t5_3, t5_2)
print(d // (24 * 3600))

print(os.getenv "HOME")