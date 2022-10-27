package main

import (
    "fmt"
    "io/ioutil"
    "os"

    "luago/binchunk"
    "luago/state"
)

import . "luago/api"
import . "luago/vm"

func main() {
    if len(os.Args) > 1 {
        data, err := ioutil.ReadFile(os.Args[1])
        if err != nil {
            panic(err)
        }

        proto := binchunk.Undump(data)
        luaMain(proto)
    }
}

func luaMain(proto *binchunk.Prototype) {
    nRegs := int(proto.MaxStackSize)
    L := state.New(nRegs + 20, proto)
    L.SetTop(nRegs)

    for {
        pc := L.PC()
        inst := Instruction(L.Fetch())
        if inst.Opcode() != OP_RETURN {
            inst.Execute(L)
            fmt.Printf("[%02d] %s ", pc + 1, inst.OpName())
            printStack(L)
        } else {
            break
        }
    }
}

func printStack(L LuaState) {
    top := L.GetTop()
    for i := 1; i <= top; i++ {
        t := L.Type(i)
        switch t {
        case LUA_TBOOLEAN: fmt.Printf("[%t]", L.ToBoolean(i))
        case LUA_TNUMBER: fmt.Printf("[%g]", L.ToNumber(i))
        case LUA_TSTRING: fmt.Printf("[%q]", L.ToString(i))
        default: fmt.Printf("[%s]", L.TypeName(t))
        }
    }

    fmt.Println()
}