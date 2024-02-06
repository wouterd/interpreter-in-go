package serializer

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func TestSerializingArray(t *testing.T) {

	tests := []struct {
		input    *object.Array
		expected []byte
	}{
		{
			input:    &object.Array{Elements: []object.Object{}},
			expected: []byte{1, 0, 0, 0, 0},
		},
		{
			input: &object.Array{Elements: []object.Object{
				&object.Integer{Value: 69420},
				&object.Boolean{Value: true},
				&object.Boolean{Value: false},
				&object.Null{},
				&object.String{Value: "hello world!"},
				&object.CompiledFunction{
					Instructions:  []byte{},
					NumLocals:     42,
					NumParameters: 69,
				},
			}},
			expected: flatten([][]byte{{1, 0, 0, 0, 6},
				binary.BigEndian.AppendUint64([]byte{2}, 69420),
				{3},
				{4},
				{5},
				{6, 'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', '!', 0},
				{7, 42, 69, 0, 0, 0, 0},
			}),
		},
	}

	for ii, tt := range tests {
		ser := New()
		err := ser.writeObj(tt.input)
		if err != nil {
			t.Fatalf("[%d]: Error writing obj: %s", ii, err.Error())
		}
		if len(ser.Output) != len(tt.expected) {
			t.Fatalf("[%d]: output length not correct. got=%d, want=%d.", ii, len(ser.Output), len(tt.expected))
		}
		for i := range tt.expected {
			if ser.Output[i] != tt.expected[i] {
				t.Fatalf("[%d]: byte [%d] not right. got=%d, expected=%d", ii, i, ser.Output[i], tt.expected[i])
			}
		}
	}
}

func TestSerializingProgram(t *testing.T) {
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

	// The last bytes of the output should be the instructions
	amInstr := len(c.Bytecode().Instructions)
	instrStart := len(s.Output) - amInstr
	testInstr(t, c.Bytecode().Instructions, s.Output[instrStart:])

	checkLen := binary.BigEndian.Uint32(s.Output[instrStart-4:])
	if int(checkLen) != len(c.Bytecode().Instructions) {
		t.Fatalf("Length of instructions not rendered right. got=%d, expected=%d", checkLen, len(c.Bytecode().Instructions))
	}

	amConsts := s.Output[HEADER_LEN+sha256.Size]
	if amConsts != 4 {
		t.Fatalf("Not enough constants serialized. got=%d, want=4", amConsts)
	}

	fmt.Println(s.Output)
}

func testInstr(t *testing.T, expected, actual code.Instructions) {
	ex := expected.String()
	ac := actual.String()

	if ex != ac {
		t.Fatalf("Instructions not right. got=%s, expected=%s", ac, ex)
	}
}

func flatten(arrs [][]byte) []byte {
	size := 0
	for _, arr := range arrs {
		size += len(arr)
	}
	result := make([]byte, 0, size)
	for _, arr := range arrs {
		result = append(result, arr...)
	}
	return result
}
