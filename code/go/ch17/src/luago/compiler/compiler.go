package compiler

import (
	"luago/binchunk"
	"luago/compiler/parser"
	"luago/compiler/codegen"
)

func Compile(chunk, chunkName string) *binchunk.Prototype {
	ast := parser.Parse(chunk, chunkName)
	proto := codegen.GenProto(ast)
	setSource(proto, chunkName)
	return proto
}

func setSource(proto *binchunk.Prototype, chunkName string) {
	proto.Source = chunkName

	for _, f := range proto.Protos {
		setSource(f, "")
	}
}