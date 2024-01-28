package ast

import (
    "reflect"
    "testing"
)

func TestModify(t *testing.T) {
    one := func() Expression { return &IntegerLiteral{Value: 1}}
    two := func() Expression { return &IntegerLiteral{Value: 2}}

    turnOneIntoTwo := func(node Node) Node {
        intLit, ok := node.(*IntegerLiteral)
        if !ok {
            return node
        }

        if intLit.Value != 1 {
            return node
        }

        intLit.Value = 2
        return intLit
    }

    tests := []struct {
        input Node
        expected Node
    }{
        {
            one(),
            two(),
        },
        {
            &Program{
                Statements: []Statement{
                    &ExpressionStatement{Expression: one()},
                },
            },
            &Program{
                Statements: []Statement{
                    &ExpressionStatement{Expression: two()},
                },
            },
        },
        {
            &InfixExpression{Left: one(), Operator: "+", Right: two()},
            &InfixExpression{Left: two(), Operator: "+", Right: two()},
        },
        {
            &InfixExpression{Left: two(), Operator: "+", Right: one()},
            &InfixExpression{Left: two(), Operator: "+", Right: two()},
        },
        {
            &PrefixExpression{Right: one(), Operator: "-"},
            &PrefixExpression{Right: two(), Operator: "-"},
        },
        {
            &IndexExpression{Left: one(), Index: one()},
            &IndexExpression{Left: two(), Index: two()},
        },
        {
            &IfExpression{
                Condition: one(),
                Consequence: &BlockStatement{
                    Statements: []Statement{
                        &ExpressionStatement{Expression: one()},
                    },
                },
                Alternative: &BlockStatement{
                    Statements: []Statement{
                        &ExpressionStatement{Expression: one()},
                    },
                },
            },
            &IfExpression{
                Condition: two(),
                Consequence: &BlockStatement{
                    Statements: []Statement{
                        &ExpressionStatement{Expression: two()},
                    },
                },
                Alternative: &BlockStatement{
                    Statements: []Statement{
                        &ExpressionStatement{Expression: two()},
                    },
                },
            },
        },
        {
            &ReturnStatement{ReturnValue: one()},
            &ReturnStatement{ReturnValue: two()},
        },
        {
            &LetStatement{Value: one()},
            &LetStatement{Value: two()},
        },
        {
            &FunctionLiteral{
                Parameters: []*Identifier{},
                Body: &BlockStatement{
                    Statements: []Statement{
                        &ExpressionStatement{Expression: one()},
                    },
                },
            },
            &FunctionLiteral{
                Parameters: []*Identifier{},
                Body: &BlockStatement{
                    Statements: []Statement{
                        &ExpressionStatement{Expression: two()},
                    },
                },
            },
        },
        {
            &ArrayLiteral{Elements: []Expression{one(), one()}},
            &ArrayLiteral{Elements: []Expression{two(), two()}},
        },
    }

    for _, tt := range tests {
        modified := Modify(tt.input, turnOneIntoTwo)
        
        equal := reflect.DeepEqual(modified, tt.expected)
        if !equal {
            t.Errorf("not equal. got=%#v, want=%#v", modified, tt.expected)
        }
    }
    
    hashLiteral := &HashLiteral{
        Data: []HashPair{ 
            { Key: one(), Value: one()},
            { Key: one(), Value: one()},
        },
    }

    Modify(hashLiteral, turnOneIntoTwo)

    for _, pair := range hashLiteral.Data {
        key, _ := pair.Key.(*IntegerLiteral)
        if key.Value != 2 {
            t.Errorf("value is not %d, got=%d", 2, key.Value)
        }
        val, _ := pair.Value.(*IntegerLiteral)
        if val.Value != 2 {
            t.Errorf("value is not %d, got=%d", 2, val.Value)
        }
    }
}
