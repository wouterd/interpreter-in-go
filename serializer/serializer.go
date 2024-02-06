package serializer

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"monkey/compiler"
	"monkey/object"
)

const (
	OBJ byte = iota
	ARRAY
	INTEGER
	BOOL_TRUE
	BOOL_FALSE
	NULL
	STRING
	COMPILED_FUNCTION

	InitialBufferSize = 10240

	VERSION = 1
)

var (
	HEADER     = []byte{42, 69, 'M', 'o', 'n', 'k', 'e', 'y', VERSION}
	HEADER_LEN = len(HEADER)
)

type Serializer struct {
	Output []byte
}

func New() *Serializer {
	return &Serializer{
		Output: make([]byte, 0, 10240),
	}
}

func (s *Serializer) Write(code *compiler.Bytecode) error {
	s.Output = append(s.Output, HEADER...)
	// Reserve space for the checksum
	s.Output = append(s.Output, make([]byte, sha256.Size)...)

	amConstants := len(code.Constants)
	if amConstants > 255 {
		return fmt.Errorf("Too many constants (%d), can only serialize 255 tops!", amConstants)
	}

	s.Output = append(s.Output, byte(amConstants))
	for _, obj := range code.Constants {
		s.writeObj(obj)
	}

	s.Output = binary.BigEndian.AppendUint32(s.Output, uint32(len(code.Instructions)))
	s.Output = append(s.Output, code.Instructions...)

	hash := sha256.Sum256(s.Output[HEADER_LEN+sha256.Size : len(s.Output)])
	copy(s.Output[HEADER_LEN:HEADER_LEN+sha256.Size], hash[:])

	return nil
}

func (s *Serializer) writeObj(obj object.Object) error {
	switch obj := obj.(type) {
	case *object.Array:
		// Format: ARRAY(1) SIZE(4) ..Elements
		s.Output = append(s.Output, ARRAY)
		size := uint32(len(obj.Elements))
		s.Output = binary.BigEndian.AppendUint32(s.Output, size)
		for _, el := range obj.Elements {
			err := s.writeObj(el)
			if err != nil {
				return err
			}
		}
		return nil

	case *object.Integer:
		// Format: INTEGER(1) VALUE(8)
		s.Output = append(s.Output, INTEGER)
		s.Output = binary.BigEndian.AppendUint64(s.Output, uint64(obj.Value))
		return nil

	case *object.String:
		// Format: STRING(1) CHARS(*) ZERO_BYTE(1)
		s.Output = append(s.Output, STRING)
		s.Output = append(s.Output, []byte(obj.Value)...)
		s.Output = append(s.Output, 0)
		return nil

	case *object.Boolean:
		// Format: BOOL_TRUE(1) | BOOL_FALSE(1)
		if obj.Value {
			s.Output = append(s.Output, BOOL_TRUE)
		} else {
			s.Output = append(s.Output, BOOL_FALSE)
		}
		return nil

	case *object.Null:
		// Format: NULL(1)
		s.Output = append(s.Output, NULL)
		return nil

	case *object.CompiledFunction:
		// Format: COMPILED_FUNCTION(1) NUM_LOCALS(1) NUM_PARAMS(1) INSTRUCTIONS_SIZE(4) INSTRUCTIONS(SIZE)
		s.Output = append(s.Output, COMPILED_FUNCTION)
		s.Output = append(s.Output, byte(obj.NumLocals))
		s.Output = append(s.Output, byte(obj.NumParameters))
		s.Output = binary.BigEndian.AppendUint32(s.Output, uint32(len(obj.Instructions)))
		s.Output = append(s.Output, obj.Instructions...)
		return nil

	default:
		return fmt.Errorf("Object of type [%T] can't be serialized", obj)
	}
}
