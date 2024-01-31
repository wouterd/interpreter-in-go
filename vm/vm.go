package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const (
    StackSize = 2048
    GlobalsSize = 65536
)

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
    globals []object.Object
	sp    int // Will point to the next value. top of the stack is stack[sp-1]
}

func New(bytecode *compiler.Bytecode) *VM {
    return &VM{
        instructions: bytecode.Instructions,
        constants:    bytecode.Constants,
        stack:        make([]object.Object, StackSize),
        globals: make([]object.Object, GlobalsSize),
        sp:           0,
    }
}

func (vm *VM) Recode(bytecode *compiler.Bytecode) {
    vm.sp = 0
    vm.instructions = bytecode.Instructions
    vm.constants = bytecode.Constants
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

        case code.OpSetGlobal:
            globalIdx := code.ReadUint16(vm.instructions[ip+1:])
            ip += 2
            vm.globals[globalIdx] = vm.pop()

        case code.OpGetGlobal:
            globalIdx := code.ReadUint16(vm.instructions[ip+1:])
            ip += 2
            err := vm.push(vm.globals[globalIdx])
            if err != nil {
                return err
            }

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		case code.OpPop:
			vm.pop()

		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}

		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1

		case code.OpJumpNotTruthy:
			condition := vm.pop()

			if !isTruthy(condition) {
				pos := int(code.ReadUint16(vm.instructions[ip+1:]))
				ip = pos - 1
			} else {
				ip += 2
			}

        case code.OpArray:
			amElems := int(code.ReadUint16(vm.instructions[ip+1:]))
            ip += 2

            elements := vm.buildArrayFromStack(amElems)
            err := vm.push(&object.Array{Elements: elements})
            if err != nil {
                return err
            }

        case code.OpHash:
			amElems := int(code.ReadUint16(vm.instructions[ip+1:]))
            ip += 2

            hash, err := vm.buildHashFromStack(amElems)
            if err != nil {
                return err
            }

            err = vm.push(hash)
            if err != nil {
                return err
            }

        case code.OpIndex:
            index := vm.pop()
            left := vm.pop()

            err := vm.executeIndexExpression(left, index)
            if err != nil {
                return err
            }

		}

	}

	return nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
    switch left := left.(type) {
    case *object.Array:
        return vm.executeArrayIndexExpression(left, index)
    case *object.Hash:
        return vm.executeHashIndexExpression(left, index)
    default:
        return fmt.Errorf("index operator not supported: %s", left.Type())
    }
}

func (vm *VM) executeArrayIndexExpression(left *object.Array, index object.Object) error {
    idx, ok := index.(*object.Integer)
    if !ok {
        return fmt.Errorf("Arrays can only be indexed by Integers, got=%T", index)
    }

    idxVal := idx.Value

    if idxVal < 0 || idxVal >= int64(len(left.Elements)) {
        return vm.push(Null)
    }

    return vm.push(left.Elements[idxVal])
}

func (vm *VM) executeHashIndexExpression(left *object.Hash, index object.Object) error {
    idx, ok := index.(object.Hashable)
    if !ok {
        return fmt.Errorf("unusable as hash key: %s", index.Type())
    }

    obj, ok := left.Get(idx)
    if !ok {
        return vm.push(Null)
    }
    return vm.push(obj)
}


func (vm *VM) buildHashFromStack(amElems int) (object.Object, error) {
    hash := object.NewHash()

    for i := vm.sp - amElems ; i < vm.sp ; i += 2 {
        key, ok := vm.stack[i].(object.Hashable)
        if !ok {
            return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
        }

        hash.Set(key, vm.stack[i+1])
    }

    vm.sp -= amElems

    return &hash, nil
}

func (vm *VM) buildArrayFromStack(amElems int) []object.Object {
    elements := make([]object.Object, amElems)
    for i := range vm.stack[vm.sp - amElems: vm.sp] {
        elements[i] = vm.stack[i]
        vm.stack[i] = nil
    }

    vm.sp -= amElems
    return elements
}

func isTruthy(obj object.Object) bool {
	// anything not false or Null is truthy
	return obj != False && obj != Null
}

func (vm *VM) executeMinusOperator() error {
	right := vm.pop()

	int, ok := right.(*object.Integer)
	if !ok {
		return fmt.Errorf("Minus operator only works on Integers, found %T", right)
	}

	return vm.push(&object.Integer{Value: -int.Value})
}

func (vm *VM) executeBangOperator() error {
	right := vm.pop()

	if isTruthy(right) {
		return vm.push(False)
	} else {
		return vm.push(True)
	}
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))

	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))

	default:
		return fmt.Errorf("uknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

    switch {
    case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOpration(op, left, right)
    case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
        return vm.executeBinaryStringOperation(op, left, right)
    default:
        return fmt.Errorf("unsupported type for binary operation: %s %s", leftType, rightType)
    }
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
    leftValue := left.(*object.String).Value
    rightValue := right.(*object.String).Value

    if op != code.OpAdd {
        return fmt.Errorf("Unknown string operation: %d", op)
    }

    return vm.push(&object.String{Value: leftValue + rightValue})
}

func (vm *VM) executeBinaryIntegerOpration(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	vm.sp--
	return vm.stack[vm.sp]
}
