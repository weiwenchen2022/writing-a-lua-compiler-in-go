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
        if err != nil { panic(err) }

        ls := state.New()
        ls.Register("tostring", tostring)
        ls.Register("print", print)
        ls.Register("getmetatable", getMetatable)
        ls.Register("setmetatable", setMetatable)

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
        break
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