package parser

import . "luago/compiler/lexer"
import . "luago/compiler/ast"

/* recursive descent parser */
func Parse(chunk, chunkName string) *Block {
	lexer := NewLexer(chunk, chunkName)
	block := parseBlock(lexer)
	lexer.NextTokenOfKind(TOKEN_EOF)
	return block
}