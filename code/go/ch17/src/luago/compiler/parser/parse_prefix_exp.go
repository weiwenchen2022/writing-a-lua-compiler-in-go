package parser

import . "luago/compiler/lexer"
import . "luago/compiler/ast"


// prefixexp ::= var | functioncall | `(` exp `)`
func parsePrefixExp(lexer *Lexer) Exp {
	var exp Exp

	if lexer.LookAhead() == TOKEN_IDENTIFIER {
		line, name := lexer.NextIdentifier() // Name
		exp = &NameExp{line, name,}
	} else { // `(` exp `)`
		exp = parseParensExp(lexer)
	}

	return _finishPrefixExp(lexer, exp)
}

func _finishPrefixExp(lexer *Lexer, exp Exp) Exp {
	for {
		switch lexer.LookAhead() {
		case TOKEN_SEP_LBRACK:
			lexer.NextToken() // `[`
			keyExp := parseExp(lexer) // exp
			lexer.NextTokenOfKind(TOKEN_SEP_RBRACK) // `]`
			exp = &TableAccessExp{lexer.Line(), exp, keyExp,}

		case TOKEN_SEP_DOT:
			lexer.NextToken() // `.`
			line, name := lexer.NextIdentifier() // Name
			keyExp := &StringExp{line, name,}
			exp = &TableAccessExp{line, exp, keyExp,}

		case TOKEN_SEP_COLON,
			TOKEN_SEP_LPAREN, TOKEN_SEP_LCURLY, TOKEN_STRING:
			exp = _finishFunctionCallExp(lexer, exp) // [`:` Name] args

		default:
			return exp
		}
	}

	return exp
}

func parseParensExp(lexer *Lexer) Exp {
	lexer.NextTokenOfKind(TOKEN_SEP_LPAREN) // `(`
	exp := parseExp(lexer) // exp
	lexer.NextTokenOfKind(TOKEN_SEP_RPAREN) // `)`

	switch exp.(type) {
	case *VarargExp, *FuncCallExp, *NameExp, *TableAccessExp:
		return &ParensExp{exp,}
	}

	return exp
}

// functioncall ::= prefixexp args | prefixexp `:` Name args
func _finishFunctionCallExp(lexer *Lexer, prefixExp Exp) *FuncCallExp {
	nameExp := _parseNameExp(lexer) // [`:` Name]
	line := lexer.Line()
	args := _parseArgs(lexer) // args
	lastLine := lexer.Line()
	return &FuncCallExp{line, lastLine, prefixExp, nameExp, args}
}

func _parseNameExp(lexer *Lexer) *StringExp {
	if lexer.LookAhead() == TOKEN_SEP_COLON {
		lexer.NextToken() // `:`
		line, name := lexer.NextIdentifier()
		return &StringExp{line, name,}
	}

	return nil
}

// args ::= `(` [explist] `)` | tableconstructor | LiterString
func _parseArgs(lexer *Lexer) (args []Exp) {
	switch lexer.LookAhead() {
	case TOKEN_SEP_LPAREN: // `(` explsit `)`
		lexer.NextToken()
		if lexer.LookAhead() != TOKEN_SEP_RPAREN {
			args = parseExpList(lexer)
		}

		lexer.NextTokenOfKind(TOKEN_SEP_RPAREN) // `)`

	case TOKEN_SEP_LCURLY: // `{` [fieldlist] `}`
		args = []Exp{parseTableConstructorExp(lexer),}

	default: // LiteralString
		line, str := lexer.NextTokenOfKind(TOKEN_STRING)
		args = []Exp{&StringExp{line, str,},}
	}

	return
}
