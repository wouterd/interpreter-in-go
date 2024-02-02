package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Compiler struct {
	constants []object.Object
	symbols   *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
    symbols := NewSymbolTable() 
    
    for i, v := range object.Builtins {
        symbols.DefineBuiltin(i, v.Name)
    }

	return &Compiler{
		constants:  []object.Object{},
		symbols:    symbols,
		scopes:     []CompilationScope{mainScope},
		scopeIndex: 0,
	}
}

func (c *Compiler) Reset() {
	c.scopes[c.scopeIndex].instructions = code.Instructions{}
	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{}
	c.scopes[c.scopeIndex].previousInstruction = EmittedInstruction{}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}

	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol, ok := c.symbols.Resolve(node.Name.Value)
		if !ok {
			symbol = c.symbols.Define(node.Name.Value)
		}
        if symbol.Scope == LocalScope {
            c.emit(code.OpSetLocal, symbol.Index)
        } else {
            c.emit(code.OpSetGlobal, symbol.Index)
        }

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jmpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastInstruction()
		}

		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.scopes[c.scopeIndex].instructions)
		c.changeOperand(jmpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastInstruction()
			}
		}

		afterAlternativePos := len(c.scopes[c.scopeIndex].instructions)
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))

	case *ast.ArrayLiteral:
		for _, elem := range node.Elements {
			err := c.Compile(elem)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		for _, p := range node.Data {
			err := c.Compile(p.Key)
			if err != nil {
				return err
			}
			err = c.Compile(p.Value)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Data)*2)

	case *ast.FunctionLiteral:
		c.enterScope()

		for _, p := range node.Parameters {
			c.symbols.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			lastPos := c.scopes[c.scopeIndex].lastInstruction.Position

			c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
			c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
		}

		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		numLocals := c.symbols.numDefinitions

		instructions := c.leaveScope()
		compiledFunc := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}

		c.emit(code.OpConstant, c.addConstant(compiledFunc))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.Identifier:
		symbol, ok := c.symbols.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("can't get global '%s', it's not defined.", node.Value)
		}
        c.loadSymbol(symbol)

	}

	return nil
}

func (c Compiler) loadSymbol(s Symbol) {
    switch s.Scope {
    case GlobalScope:
        c.emit(code.OpGetGlobal, s.Index)
    case LocalScope:
        c.emit(code.OpGetLocal, s.Index)
    case BuiltinScope:
        c.emit(code.OpGetBuiltin, s.Index)
    }
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	copy(c.scopes[c.scopeIndex].instructions[pos:], newInstruction)
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.scopes[c.scopeIndex].instructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	c.scopes[c.scopeIndex].previousInstruction = c.scopes[c.scopeIndex].lastInstruction
	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.scopes[c.scopeIndex].instructions) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastInstruction() {
	c.scopes[c.scopeIndex].instructions = c.scopes[c.scopeIndex].instructions[:c.scopes[c.scopeIndex].lastInstruction.Position]
	c.scopes[c.scopeIndex].lastInstruction = c.scopes[c.scopeIndex].previousInstruction
	c.scopes[c.scopeIndex].previousInstruction = EmittedInstruction{}
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.scopes[c.scopeIndex].instructions)
	c.scopes[c.scopeIndex].instructions = append(c.scopes[c.scopeIndex].instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.scopes[c.scopeIndex].instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbols = NewEnclosedSymbolTable(c.symbols)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.scopes[c.scopeIndex].instructions
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbols = c.symbols.outer
	return instructions
}
