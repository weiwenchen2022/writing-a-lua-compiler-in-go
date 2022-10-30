package stdlib

import (
	"os"
	"strings"
)

import . "luago/api"

/* key, in the registry, for table of loaded modules */
const LUA_LOADED_TABLE = "_LOADED"

/* key, in the registry, for table of preloaded loaders */
const LUA_PRELOADED_TABLE = "_PRELOAD"

const (
	LUA_DIR_SEP = string(os.PathSeparator)
	LUA_PATH_SEP = ";"
	LUA_PATH_MARK = "?"
	LUA_EXEC_DIR = "!"
	LUA_IGN_MARK = "-"
)

const LUA_PATH = "./?.lua;./?/init.lua"

var llFuncs = []FuncReg {
	{"require", pkgRequire,},
}

var pkgFuncs = []FuncReg {
	{"searchpath", pkgSearchPath,},

	/* placeholders */
	{"preload", nil,},
	{"path", nil,},
	{"searchers", nil,},
	{"loaded", nil,},
}

// require(modname)
// http://www.lua.org/manual/5.3/manual.html#pdf-require
func pkgRequire(ls LuaState) int {
	name := ls.CheckString(1)
	ls.SetTop(1) /* LOADED table will be at index 2 */
	ls.GetField(LUA_REGISTRYINDEX, LUA_LOADED_TABLE)
	ls.GetField(2, name) /* LOADED[name] */
	if ls.ToBoolean(-1) { /* is it there? */
		return 1 /* package is already loaded */
	}

	/* else must load package */
	ls.Pop(1) /* remove 'getfield' result */

	filename := _findLoader(ls, name)
	ls.PushString(name)
	ls.Insert(-2) /* name is 1st arguments (before search data) */
	ls.Call(2, 1) /* run loader to load module */
	if !ls.IsNil(-1) { /* non-nil return? */
		ls.PushValue(-1)
		ls.SetField(2, name) /* LOADED[name] = returned value */
	}

	if LUA_TNIL == ls.GetField(2, name) { /* module set no value? */
		ls.PushBoolean(true) /* use true as result */
		ls.PushValue(-1) /* extra copy to be returned */
		ls.SetField(2, name) /* LOADED[name] = true */
	}

	ls.PushString(filename)
	return 2
}

func _findLoader(ls LuaState, name string) string {
	/* push 'package.searchers' to index 3 in the stack */
	if ls.GetField(LuaUpvalueIndex(1), "searchers") != LUA_TTABLE {
		ls.Error2("'package.searchers' must be a table")
	}

	/* to build error message */
	errMsg := "module '" + name + "' not found:"

	/* iterate over available searchers to find a loader */
	for i := 1; ; i++ {
		if LUA_TNIL == ls.RawGetI(3, int64(i)) { /* no more searchers? */
			ls.Pop(1) /* remove nil */
			ls.Error2(errMsg) /* create error message */
		}

		ls.PushString(name)
		ls.Call(1, 2) /* call it */
		if ls.IsFunction(-2) { /* did it find a loader? */
			return ls.ToString(-1) /* module loader found */
		} else if ls.IsString(-2) { /* searcher returned error message? */
			ls.Pop(1) /* remove extra return */
			errMsg += ls.ToString(-1) /* concatenate error message */
		} else {
			ls.Pop(2) /* remove both returns */
		}
	}
}

// package.searchpath(name, path [, sep [, rep]])
// http://www.lua.org/manual/5.3/manual.html#pdf-package.searchpath
// loadlib.c#ll_searchpath
func pkgSearchPath(ls LuaState) int {
	name := ls.CheckString(1)
	path := ls.CheckString(2)
	sep := ls.OptString(3, ".")
	rep := ls.OptString(4, LUA_DIR_SEP)

	if filename, errMsg := _searchPath(name, path, sep, rep); errMsg == "" {
		ls.PushString(filename)
		return 1
	} else {
		ls.PushNil()
		ls.PushString(errMsg)
		return 2
	}
}

func _searchPath(name, path, sep, rep string) (filename, errMsg string) {
	if sep != "" {
		name = strings.Replace(name, sep, rep, -1)
	}

	for _, filename := range strings.Split(path, LUA_PATH_SEP) {
		filename = strings.Replace(filename, LUA_PATH_MARK, name, -1)
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			return filename, ""
		}

		errMsg += "\n\tno file '" + filename + "'"
	}

	return "", errMsg
}

func createSearchersTable(ls LuaState) {
	searchers := []GoFunction {
		preloadSearcher,
		luaSearcher,
	}

	/* create 'searchers' table */
	ls.CreateTable(len(searchers), 0)

	/* fill it with predefined searchers */
	for i, searcher := range searchers {
		ls.PushValue(-2) /* set 'package' as upvalue for all searchers */
		ls.PushGoClosure(searcher, 1)
		ls.RawSetI(-2, int64(i + 1))
	}

	ls.SetField(-2, "searchers") /* put it in field 'searchers' */
}

func preloadSearcher(ls LuaState) int {
	name := ls.CheckString(1)

	ls.GetField(LUA_REGISTRYINDEX, "_PRELOAD")
	if LUA_TNIL == ls.GetField(-1, name) { /* not found? */
		ls.PushString("\n\tno field package.preload['" + name + "']")
	} else {
		ls.PushString(":preload:")
	}

	return 2
}

func luaSearcher(ls LuaState) int {
	name := ls.CheckString(1)
	ls.GetField(LuaUpvalueIndex(1), "path")
	path, ok := ls.ToStringX(-1)
	if !ok {
		ls.Error2("'package.path' must be a string")
	}

	filename, errMsg := _searchPath(name, path, ".", LUA_DIR_SEP)
	if errMsg != "" {
		ls.PushString(errMsg);
		return 1
	}

	if LUA_OK == ls.LoadFile(filename) { /* module loaded successfully? */
		ls.PushString(filename) /* will be 2nd argument to module */
		return 2 /* return loader function and filename */
	} else {
		return ls.Error2("error loading module '%s' from file '%s':\n\t%s",
			name, filename, ls.ToString(-1))
	}
}

func OpenPackageLib(ls LuaState) int {
	ls.NewLib(pkgFuncs) /* create 'package' table */
	createSearchersTable(ls)

	/* set paths */
	var path string
	if path = os.Getenv("LUA_PATH_5_3"); path != "" {

	} else if path = os.Getenv("LUA_PATH"); path != "" {

	} else {
		path = LUA_PATH
	}

	path = strings.Replace(path, ";;", ";" + LUA_PATH, -1)
	ls.PushString(path)
	ls.SetField(-2, "path")

	/* store config information */
	ls.PushString(LUA_DIR_SEP + "\n" +
		LUA_PATH_SEP + "\n" +
		LUA_PATH_MARK + "\n" +
		LUA_EXEC_DIR + "\n" +
		LUA_IGN_MARK + "\n")
	ls.SetField(-2, "config")

	/* set field 'loaded' */
	ls.GetSubTable(LUA_REGISTRYINDEX, LUA_LOADED_TABLE)
	ls.SetField(-2, "loaded")

	/* set field 'preload' */
	ls.GetSubTable(LUA_REGISTRYINDEX, LUA_PRELOADED_TABLE)
	ls.SetField(-2, "preload")

	ls.PushGlobalTable()
	ls.PushValue(-2) /* set 'package' as upvalue for next lib */
	ls.SetFuncs(llFuncs, 1) /* open lib into global table */
	ls.Pop(1) /* pop global table */
	return 1 /* return 'package' table */
}