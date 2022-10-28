package main

import (
    "fmt"
    "io/ioutil"
    "os"
)

import . "luago/compiler/lexer"

func main() {
    if len(os.Args) > 1 {
        data, err := ioutil.ReadFile(os.Args[1])
        if err != nil {
            panic(err)
        }

        testLexer(string(data), os.Args[1])
    }
}

func testLexer(chunk, chunkName string) {
    lexer := NewLexer(chunk, chunkName)
    for {
        line, kind, token := lexer.NextToken()
        fmt.Printf("[%2d] [%-10s] %s\n",
            line, kindToCategory(kind), token)
        if TOKEN_EOF == kind {
            break
        }
    }
}

func kindToCategory(kind int) string {
    switch {
    case kind < TOKEN_SEP_SEMI:
        return "other"
    
    case kind <= TOKEN_SEP_RCURLY:
        return "separator"
    
    case kind <= TOKEN_OP_NOT:
        return "operator"
    
    case kind <= TOKEN_KW_WHILE:
        return "keyword"
    
    case TOKEN_IDENTIFIER == kind:
        return "identifier"
    
    case TOKEN_NUMBER == kind:
        return "number"
    
    case TOKEN_STRING == kind:
        return "string"

    default:
        return "other"
    }
}