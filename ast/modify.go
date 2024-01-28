package ast

type ModifierFunc func(Node) Node

func Modify(node Node, modifier ModifierFunc) Node {
    mod := func(node Node) Node {
        return Modify(node, modifier)
    }

    if node == nil {
        return node
    }
    
    switch node := node.(type) {
    case *Program:
        for i, stmt := range node.Statements {
            node.Statements[i], _ = mod(stmt).(Statement)
        }
        
    case *BlockStatement:
        for i, stmt := range node.Statements {
            node.Statements[i], _ = mod(stmt).(Statement)
        }
        
    case *ExpressionStatement:
        node.Expression, _ = mod(node.Expression).(Expression)

    case *ReturnStatement:
        node.ReturnValue, _ = mod(node.ReturnValue).(Expression)

    case *LetStatement:
        node.Value, _ = mod(node.Value).(Expression)

    case *InfixExpression:
        node.Left, _ = mod(node.Left).(Expression)
        node.Right, _ = mod(node.Right).(Expression)

    case *PrefixExpression:
        node.Right, _ = mod(node.Right).(Expression)
        
    case *IndexExpression:
        node.Left, _ = mod(node.Left).(Expression)
        node.Index, _ = mod(node.Index).(Expression)
        
    case *IfExpression:
        node.Condition, _ = mod(node.Condition).(Expression)
        node.Consequence, _ = mod(node.Consequence).(*BlockStatement)
        node.Alternative, _ = mod(node.Alternative).(*BlockStatement)

    case *FunctionLiteral:
        node.Body, _ = mod(node.Body).(*BlockStatement)
        for i, param := range node.Parameters {
            node.Parameters[i] = mod(param).(*Identifier)
        }

    case *ArrayLiteral:
        for i, exp := range node.Elements {
            node.Elements[i] = mod(exp).(Expression)
        }

    case *HashLiteral:
        for _, pair := range node.Data {
            pair.Key = mod(pair.Key).(Expression)
            pair.Value = mod(pair.Value).(Expression)
        }

    }

    return modifier(node)
}
