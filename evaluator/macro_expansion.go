package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

func DefineMacros(program *ast.Program, env *object.Environment) {
	macroIdxs := []int{}

	for i, stmt := range program.Statements {
		if isMacroDefinition(stmt) {
			macroIdxs = append(macroIdxs, i)
			addMacro(env, stmt)
		}
	}

	if len(macroIdxs) == 0 {
		return
	}

	prevI := -1
	newStmts := make([]ast.Statement, 0, len(program.Statements)-len(macroIdxs))
	for i := 0; i < len(macroIdxs); i++ {
		if prevI > 0 {
			newStmts = append(newStmts, program.Statements[prevI+1:macroIdxs[i]]...)
		} else {
			newStmts = append(newStmts, program.Statements[:macroIdxs[i]]...)
		}
		prevI = macroIdxs[i]
	}

	newStmts = append(newStmts, program.Statements[prevI+1:]...)
	program.Statements = newStmts
}

func ExpandMacros(program ast.Node, env *object.Environment) ast.Node {
	return ast.Modify(program, func(n ast.Node) ast.Node {
		call, ok := n.(*ast.CallExpression)
		if !ok {
			return n
		}

		macro, ok := isMacroCall(call, env)
		if !ok {
			return n
		}

		args := quoteArgs(call)
		evalEnv := extendMacroEnv(macro, args)

		evaluated := Eval(macro.Body, evalEnv)

		quote, ok := evaluated.(*object.Quote)

		if !ok {
			panic("we only support returning AST-nodes from macros")
		}

		return quote.Node
	})
}

func extendMacroEnv(macro *object.Macro, args []*object.Quote) *object.Environment {
	extended := object.NewEnclosedEnvironment(macro.Env)

	for i, param := range macro.Parameters {
		extended.Set(param.Value, args[i])
	}

	return extended
}

func quoteArgs(call *ast.CallExpression) []*object.Quote {
	args := make([]*object.Quote, 0, len(call.Arguments))

	for _, a := range call.Arguments {
		args = append(args, &object.Quote{Node: a})
	}

	return args
}

func isMacroCall(call *ast.CallExpression, env *object.Environment) (*object.Macro, bool) {
	id, ok := call.Function.(*ast.Identifier)
	if !ok {
		return nil, false
	}

	obj, ok := env.Get(id.Value)
	if !ok {
		return nil, false
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		return nil, false
	}

	return macro, true
}

func isMacroDefinition(node ast.Statement) bool {
	letStmt, ok := node.(*ast.LetStatement)
	if !ok {
		return false
	}

	_, isMacro := letStmt.Value.(*ast.MacroLiteral)

	return isMacro
}

func addMacro(env *object.Environment, stmt ast.Statement) {
	letStmt, _ := stmt.(*ast.LetStatement)
	macroLit, _ := letStmt.Value.(*ast.MacroLiteral)

	macro := &object.Macro{
		Parameters: macroLit.Parameters,
		Env:        env,
		Body:       macroLit.Body,
	}

	env.Set(letStmt.Name.Value, macro)
}
