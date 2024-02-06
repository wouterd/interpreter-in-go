package serializer

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"monkey/compiler"
	"monkey/object"
)

type Loader struct {
	input     []byte
	len       int
	pos       int
	constants []object.Object
	used      bool
}

func NewLoader(buf []byte) *Loader {
	return &Loader{
		input:     buf,
		len:       len(buf),
		pos:       0,
		constants: make([]object.Object, 255),
		used:      false,
	}
}

func (l *Loader) Load() (*compiler.Bytecode, error) {
	if l.used {
		return nil, fmt.Errorf("This loader has already been used, create a new one")
	}
	l.used = true

	if err := l.checkHeader(); err != nil {
		return nil, err
	}

	amConsts := l.input[l.pos]
	l.pos++

	for i := 0; i < int(amConsts); i++ {
		constant, err := l.readConstant()
		if err != nil {
			return nil, fmt.Errorf("Error reading constant #%d: %s", i, err.Error())
		}
		l.constants[i] = constant
	}

	instrLen := binary.BigEndian.Uint32(l.input[l.pos : l.pos+4])
	l.pos += 4

	instr := make([]byte, instrLen)
	copy(instr, l.input[l.pos:l.pos+int(instrLen)])

	return &compiler.Bytecode{
		Constants:    l.constants[:amConsts],
		Instructions: instr,
	}, nil
}

func (l *Loader) readConstant() (object.Object, error) {
	if l.pos >= l.len {
		return nil, fmt.Errorf("Can't read type byte, no more data in buffer")
	}
	cType := l.input[l.pos]
	l.pos += 1

	switch cType {
	case STRING:
		return l.readString()

	case INTEGER:
		return l.readInteger()

	case COMPILED_FUNCTION:
		return l.readFunction()

	default:
		return nil, fmt.Errorf("Can't load constant type value %d.", cType)
	}
}

func (l *Loader) readString() (*object.String, error) {
	left := l.pos
	for l.pos < l.len && l.input[l.pos] != 0 {
		l.pos++
	}
	if l.pos >= l.len {
		return nil, fmt.Errorf("No string-terminating 0-byte found.")
	}
	str := string(l.input[left:l.pos])
	l.pos++
	return &object.String{Value: str}, nil
}

func (l *Loader) readInteger() (*object.Integer, error) {
	if l.pos+8 > l.len {
		return nil, fmt.Errorf("not enough data in buffer to read INTEGER")
	}
	val := binary.BigEndian.Uint64(l.input[l.pos:])
	l.pos += 8
	return &object.Integer{Value: int64(val)}, nil
}

func (l *Loader) readFunction() (*object.CompiledFunction, error) {
	if l.pos+6 > l.len {
		return nil, fmt.Errorf("Can't read function header, not enough data in buffer")
	}

	cf := &object.CompiledFunction{
		NumLocals:     int(l.input[l.pos]),
		NumParameters: int(l.input[l.pos+1]),
	}

	instrLen := binary.BigEndian.Uint32(l.input[l.pos+2:])
	l.pos += 6

	if l.pos+int(instrLen) > l.len {
		return nil, fmt.Errorf("Can't read function instructions. Not %d bytes left in buffer", instrLen)
	}

	cf.Instructions = make([]byte, instrLen)
	copy(cf.Instructions, l.input[l.pos:l.pos+int(instrLen)])

	l.pos += int(instrLen)

	return cf, nil
}

func (l *Loader) checkHeader() error {
	for i, b := range HEADER {
		if l.input[l.pos+i] != b {
			return fmt.Errorf("error in file header at byte #%d", i+1)
		}
	}
	l.pos = len(HEADER)

	checkHash := l.input[l.pos : l.pos+sha256.Size]
	l.pos += sha256.Size
	if len(checkHash) != sha256.Size {
		return fmt.Errorf("Not enough data in the buffer for the hash")
	}

	hash := sha256.Sum256(l.input[l.pos:])
	for i, b := range hash {
		if checkHash[i] != b {
			return fmt.Errorf("Byte #%d of checksum doesn't match.", i+1)
		}
	}

	return nil
}
