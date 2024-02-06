package serializer

import (
	"monkey/compiler"
	"monkey/lexer"
	"monkey/parser"
	"testing"
)

func TestSerialzeAndLoad(t *testing.T) {
	input := `
        let a = 1;
        let calc = fn(x) {
            return x + a
        }
        calc(42)
        puts("Hello Monkey!")
    `

	l := lexer.New(input)
	p := parser.New(l)
	c := compiler.New()
	c.Compile(p.ParseProgram())

	s := New()
	s.Write(c.Bytecode())

	loader := NewLoader(s.Output)
	bytecode, err := loader.Load()

	if err != nil {
		t.Fatalf("Loader had an error: %s", err.Error())
	}

	actual := bytecode.Instructions.String()
	expected := c.Bytecode().Instructions.String()
	if actual != expected {
		t.Fatalf("Instructions don't match, got=%s, expected=%s", actual, expected)
	}
}
