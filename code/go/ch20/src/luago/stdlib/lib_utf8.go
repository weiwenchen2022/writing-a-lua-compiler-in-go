package stdlib

import (
	"unicode/utf8"
)

import . "luago/api"

var utf8Lib = []FuncReg {
	{"len", utfLen,},
	{"offset", utfByteOffset,},
	// {"codepoint", utfCodePoint,},
	{"char", utfChar,},
	// {"codes", utfIterCodes,},
}

// utf8.len(s [, i [, j]])
// http://www.lua.org/manual/5.3/manual.html#pdf-utf8.len
// lua-5.3.5/src/lutf8lib.c#utflen()
func utfLen(ls LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	i := posRelat(ls.OptInteger(2, 1), sLen)
	j := posRelat(ls.OptInteger(3, -1), sLen)

	ls.ArgCheck(1 <= i && i <= sLen + 1, 2,
		"initial positon out of string")
	ls.ArgCheck(j <= sLen, 3,
		"final postion out of string")

	if i > j {
		ls.PushInteger(0)
	} else {
		n := utf8.RuneCountInString(s[i - 1 : j])
		ls.PushInteger(int64(n))
	}

	return 1
}

// utf8.offset(s, n [, i])
// http://www.lua.org/manual/5.3/manual.html#pdf-utf8.offset
func utfByteOffset(ls LuaState) int {
	panic("todo!")
}

// utf8.char(···)
// http://www.lua.org/manual/5.3/manual.html#pdf-utf8.char
// lua-5.3.5/src/lutf8lib.c#utfchar()
func utfChar(ls LuaState) int {
	n := ls.GetTop() /* number of arguments */
	codePoints := make([]rune, n)

	for i := 1; i <= n; i++ {
		cp := ls.CheckInteger(i)
		ls.ArgCheck(0 <= cp && cp <= 0x10FFFF, i,
			"value out of range")
		codePoints[i - 1] = rune(cp)
	}

	ls.PushString(_encodeUtf8(codePoints))
	return 1
}

func _encodeUtf8(codePoints []rune) string {
	buf := make([]byte, 6)
	str := make([]byte, 0, len(codePoints))

	for _, cp := range codePoints {
		n := utf8.EncodeRune(buf, cp)
		str = append(str, buf[:n]...)
	}

	return string(str)
}

func OpenUTF8Lib(ls LuaState) int {
	ls.NewLib(utf8Lib)
	return 1
}