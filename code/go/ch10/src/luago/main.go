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

func tostring(L LuaState) int {
    L.SetTop(1)

    t := L.Type(1)
    switch t {
    case LUA_TBOOLEAN:
        if L.ToBoolean(1) {
            L.PushString("true")
        } else {
            L.PushString("false")
        }

        break

    case LUA_TNUMBER:
    case LUA_TSTRING:
        L.ToString(1)
        break

    default:
        L.PushString(L.TypeName(t))
        break
    }

    return 1;
}

func print(L LuaState) int {
    nArgs := L.GetTop()
    L.GetGlobal("tostring")

    for i := 1; i <= nArgs; i++ {
        L.PushValue(-1)
        L.PushValue(i)
        L.Call(1, 1)

        if i > 1 {
            fmt.Print("\t")
        }
        
        fmt.Print(L.ToString(-1))
        L.Pop(1)
        // if L.IsBoolean(i) {
        //     fmt.Printf("%t", L.ToBoolean(i))
        // } else if L.IsString(i) {
        //     fmt.Print(L.ToString(i))
        // } else {
        //     fmt.Print(L.TypeName(L.Type(i)))
        // }

        // if i < nArgs {
        //     fmt.Print("\t")
        // }
    }

    fmt.Println()
    return 0
}