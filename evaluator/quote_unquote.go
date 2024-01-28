package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
	"monkey/token"
)

func quote(node ast.Node, env *object.Environment) object.Object {
	node = evalUnquoteCalls(node, env)
	return &object.Quote{Node: node}
}

func evalUnquoteCalls(quoted ast.Node, env *object.Environment) ast.Node {
	return ast.Modify(quoted, func(n ast.Node) ast.Node {
		if !isUnquoteCall(n) {
			return n
		}

		call, ok := n.(*ast.CallExpression)
		if !ok {
			return n
		}

		if len(call.Arguments) != 1 {
			return n
		}

		unquoted := Eval(call.Arguments[0], env)
		return convertToAstNode(unquoted)
	})
}

func isUnquoteCall(node ast.Node) bool {
	callExp, ok := node.(*ast.CallExpression)
	return ok && callExp.Function.TokenLiteral() == "unquote"
}

func convertToAstNode(obj object.Object) ast.Node {
	switch obj := obj.(type) {
	case *object.Integer:
		t := token.Token{
			Type:    token.INT,
			Literal: fmt.Sprintf("%d", obj.Value),
		}
		return &ast.IntegerLiteral{Token: t, Value: obj.Value}

	case *object.Boolean:
		if obj.Value {
			return &ast.Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}}
		} else {
			return &ast.Boolean{Token: token.Token{Type: token.FALSE, Literal: "false"}}
		}

	case *object.Quote:
		return obj.Node

	default:
		return nil
	}
}
