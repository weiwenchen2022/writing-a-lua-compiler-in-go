package main

import (
    "fmt"
    "io/ioutil"
    "os"

    "luago/state"
)

import . "luago/api"

func main() {
    if len(os.Args) > 1 {
        data, err := ioutil.ReadFile(os.Args[1])
        if err != nil {
            panic(err)
        }

        ls := state.New()
        ls.Register("tostring", tostring)
        ls.Register("print", print)

        ls.Register("getmetatable", getMetatable)
        ls.Register("setmetatable", setMetatable)

        ls.Register("next", next)
        ls.Register("pairs", pairs)
        ls.Register("ipairs", iPairs)

        ls.Register("error", error)
        ls.Register("pcall", pCall)

        ls.Load(data, os.Args[1], "b")
        ls.Call(0, 0)
    }
}

func tostring(ls LuaState) int {
    ls.SetTop(1)

    t := ls.Type(1)
    switch t {
    case LUA_TBOOLEAN:
        ls.PushString(fmt.Sprintf("%t", ls.ToBoolean(1)))

    case LUA_TNUMBER, LUA_TSTRING:
        ls.ToString(1)

    default:
        ls.PushString(ls.TypeName(t))
    }

    return 1;
}

func print(ls LuaState) int {
    nArgs := ls.GetTop()
    ls.GetGlobal("tostring")

    for i := 1; i <= nArgs; i++ {
        ls.PushValue(-1)
        ls.PushValue(i)
        ls.Call(1, 1)

        if i > 1 {
            fmt.Print("\t")
        }
        
        fmt.Print(ls.ToString(-1))
        ls.Pop(1)
    }

    fmt.Println()
    return 0
}

func getMetatable(ls LuaState) int {
    if !ls.GetMetatable(1) {
        ls.PushNil()
    }

    return 1
}

func setMetatable(ls LuaState) int {
    ls.SetMetatable(1)
    return 1
}

func next(ls LuaState) int {
    ls.SetTop(2)
    if ls.Next(1) {
        return 2
    } else {
        ls.PushNil()
        return 1
    }
}

func pairs(ls LuaState) int {
    ls.PushGoFunction(next) /* will return generator, */
    ls.PushValue(1) /* state, */
    ls.PushNil()
    return 3
}

func iPairs(ls LuaState) int {
    ls.PushGoFunction(_iPairsAux) /* iteration function */
    ls.PushValue(1) /* state */
    ls.PushInteger(0) /* initial value */
    return 3
}

func _iPairsAux(ls LuaState) int {
    i := ls.ToInteger(2) + 1
    ls.PushInteger(i)

    if LUA_TNIL == ls.GetI(1, i) {
        return 1
    } else {
        return 2
    }
}

func error(ls LuaState) int {
    return ls.Error()
}

func pCall(ls LuaState) int {
    nArgs := ls.GetTop() - 1
    status := ls.PCall(nArgs, -1, 0)
    ls.PushBoolean(LUA_OK == status)
    ls.Insert(1)
    return ls.GetTop()
}