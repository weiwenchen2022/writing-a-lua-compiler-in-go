
print(os.time())

local date = 1439653520
local day2year = 365.242
local sec2hour = 60 * 60
local sec2day = sec2hour * 24
local sec2year = sec2day * day2year

-- year
print(date // sec2year + 1970)

-- hour
print(date % sec2day // sec2hour)

-- minute
print(date % sec2hour // 60)

-- second
print(date % 60)

print(os.time {year = 2015, month = 8, day = 15, hour = 12, min = 45, sec = 20,})
print(os.time {year = 1970, month = 1, day = 1, hour = 0,})
print(os.time {year = 1970, month = 1, day = 1, hour = 0, sec = 1,})
print(os.time {year = 1970, month = 1, day = 1,})

local t5_3 = os.time {year = 2015, month = 1, day = 12,}
local t5_2 = os.time {year = 2011, month = 12, day = 16,}
local d = os.difftime(t5_3, t5_2)
print(d // (24 * 3600))

myepoch = os.time {year = 2000, month = 1, day = 1, hour = 0,}
now = os.time {year = 2015, month = 11, day = 20,}
print(os.difftime(now, myepoch))