package main

import (
    "encoding/json"
    "io/ioutil"
    "os"

    "luago/compiler/parser"
)

func main() {
    if len(os.Args) > 1 {
        data, err := ioutil.ReadFile(os.Args[1])
        if err != nil { panic(err) }

        testParser(string(data), os.Args[1])
    }
}

func testParser(chunk, chunkName string) {
    ast := parser.Parse(chunk, chunkName)
    b, err := json.Marshal(ast)
    if err != nil {
        panic(err)
    }

    println(string(b))
}