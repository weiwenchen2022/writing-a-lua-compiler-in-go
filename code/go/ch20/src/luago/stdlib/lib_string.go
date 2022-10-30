package stdlib

import (
	"fmt"
	"strings"
)

import . "luago/api"

var strLib = []FuncReg {
	{"len", strLen,}, // string.len(s)
	
	{"rep", strRep,}, // string.rep(s, n [, sep])
	
	{"reverse", strReverse,}, // string.reverse(s)
	
	{"lower", strLower,}, // string.lower(s)
	{"upper", strUpper,}, // string.upper(s)

	{"sub", strSub,}, // string.sub(s, i [, j])
	
	{"byte", strByte,}, // string.byte(s [, i [, j]])
	{"char", strChar,}, // string.char(...)
	
	{"format", strFormat,}, // string.format(formatstring, ...)
}

/* Basic String Functions */

// string.len(s)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.len
// lua-5.3.5/src/lstrlib.c#str_len()
func strLen(ls LuaState) int {
	s := ls.CheckString(1)
	ls.PushInteger(int64(len(s)))
	return 1
}

// string.rep(s, n [, sep])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.rep
// lua-5.3.5/src/lstrlib.c#str_rep()
func strRep(ls LuaState) int {
	s := ls.CheckString(1)
	n := ls.CheckInteger(2)
	sep := ls.OptString(3, "")

	if n <= 0 {
		ls.PushString("")
	} else if n == 1 {
		ls.PushString(s)
	} else {
		a := make([]string, n)
		for i := 0; i < int(n); i++ {
			a[i] = s
		}

		ls.PushString(strings.Join(a, sep))
	}

	
	return 1
}

// string.reverse(s)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.reverse
// lua-5.3.5/src/lstrlib.c#str_reverse()
func strReverse(ls LuaState) int {
	s := ls.CheckString(1)

	if strLen := len(s); strLen > 1 {
		a := make([]byte, strLen)

		for i := 0; i < strLen; i++ {
			a[i] = s[strLen - i - 1]
		}

		ls.PushString(string(a))
	}

	return 1
}

// string.lower(s)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.lower
// lua-5.3.5/src/lstrlib.c#str_lower()
func strLower(ls LuaState) int {
	s := ls.CheckString(1)
	ls.PushString(strings.ToLower(s))
	return 1
}

// string.upper(s)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.upper
// lua-5.3.5/src/lstrlib.c#str_upper()
func strUpper(ls LuaState) int {
	s := ls.CheckString(1)
	ls.PushString(strings.ToUpper(s))
	return 1
}

// string.sub(s, i [, j])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.sub
// lua-5.3.5/src/lstrlib.c#str_sub()
func strSub(ls LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	i := posRelat(ls.CheckInteger(2), sLen)
	j := posRelat(ls.OptInteger(3, -1), sLen)

	if i < 1 {
		i = 1
	}

	if j > sLen {
		j = sLen
	}

	if i <= j {
		ls.PushString(s[i-1 : j])
	} else {
		ls.PushString("")
	}
	
	return 1
}

// string.byte(s [, i [, j]])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.byte
// lua-5.3.5/src/lstrlib.c#str_byte()
func strByte(ls LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	i := posRelat(ls.OptInteger(2, 1), sLen)
	j := posRelat(ls.OptInteger(3, int64(i)), sLen)

	if i < 1 {
		i = 1
	}

	if j > sLen {
		j = sLen
	}

	if i > j { /* empty interval; return no values */
		return 0
	}

	n := j - i + 1
	ls.CheckStack2(n, "string slice too long")

	for ; i <= j; i++ {
		ls.PushInteger(int64(s[i-1]))
	}

	return n
}

// string.char(···)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.char
// lua-5.3.5/src/lstrlib.c#str_char()
func strChar(ls LuaState) int {
	nArgs := ls.GetTop()

	if nArgs == 0 {
		ls.PushString("")
	} else {
		s := make([]byte, nArgs)

		for i := 1; i <= nArgs; i++ {
			c := ls.CheckInteger(i)
			ls.ArgCheck(c == int64(byte(c)), i, "value out of range")
			s[i - 1] = byte(c)
		}

		ls.PushString(string(s))
	}

	return 1
}

/* STRING FORMAT */

// string.format(formatstring, ···)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.format
func strFormat(ls LuaState) int {
	fmtStr := ls.CheckString(1)

	if len(fmtStr) <= 1 || strings.IndexByte(fmtStr, '%') < 0 {
		ls.PushString(fmtStr)
		return 1
	}

	argIdx := 1
	arr := parseFmtStr(fmtStr)
	for i, s := range arr {
		if s[0] == '%' {
			if s == "%%" {
				arr[i] = "%"
			} else {
				argIdx += 1
				arr[i] = _fmtArg(s, ls, argIdx)
			}
		}
	}

	ls.PushString(strings.Join(arr, ""))
	return 1
}

func _fmtArg(tag string, ls LuaState, argIdx int) string {
	switch tag[len(tag) - 1] { // specifier
	case 'c': // character
		return string([]byte{byte(ls.ToInteger(argIdx)),})

	case 'i':
		tag = tag[:len(tag) - 1] + "%d" // %i -> %d
		return fmt.Sprintf(tag, ls.ToInteger(argIdx))

	case 'd', 'o': // integer, octal
		return fmt.Sprintf(tag, ls.ToInteger(argIdx))

	case 'u': // unsigned integer
		tag = tag[: len(tag) - 1] + "d" // %u -> %d
		return fmt.Sprintf(tag, uint(ls.ToInteger(argIdx)))

	case 'x', 'X': // hex integer
		return fmt.Sprintf(tag, uint(ls.ToInteger(argIdx)))

	case 'f': // float
		return fmt.Sprintf(tag, ls.ToNumber(argIdx))

	case 's', 'q': // string
		return fmt.Sprintf(tag, ls.ToString2(argIdx))

	default:
		panic("todo! tag=" + tag)
	}
}
/* helper */

/* translate a relative string position: negative means back from end */
func posRelat(pos int64, _len int) int {
	_pos := int(pos)

	if _pos >= 0 {
		return _pos
	} else if -_pos > _len {
		return 0
	} else {
		return _len + _pos + 1
	}
}

func OpenStringLib(ls LuaState) int {
	ls.NewLib(strLib)
	createMetatable(ls)
	return 1
}

func createMetatable(ls LuaState) {
	ls.CreateTable(0, 1) /* table to be metatable for strings */
	ls.PushString("dummy") /* dummy string */
	ls.PushValue(-2) /* copy table */
	ls.SetMetatable(-2) /* set table as metateble for strings */
	ls.Pop(1) /* pop dummy string */
	ls.PushValue(-2) /* get string library */
	ls.SetField(-2, "__index") /* metatable.__index = string */
	ls.Pop(1) /* pop metatable */
}