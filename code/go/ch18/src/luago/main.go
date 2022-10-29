package main

import (
    "os"

    "luago/state"
)

import . "luago/api"

func main() {
    if len(os.Args) > 1 {
        ls := state.New()
        ls.OpenLibs()
        ls.LoadFile(os.Args[1])
        ls.Call(0, LUA_MULTRET)
    }
}