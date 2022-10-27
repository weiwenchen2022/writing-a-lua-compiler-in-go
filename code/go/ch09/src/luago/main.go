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

        L := state.New()
        L.Register("tostring", tostring)
        L.Register("print", print)

        L.Load(data, os.Args[1], "b")
        L.Call(0, 0)
    }
}

func tostring(ls LuaState) int {
    ls.SetTop(1)

    t := ls.Type(1)
    switch t {
    case LUA_TBOOLEAN:
        ls.PushString(fmt.Sprintf("%t", ls.ToBoolean))
        break

    case LUA_TSTRING, LUA_TNUMBER:
        ls.ToString(1)
        break

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