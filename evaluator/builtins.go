package evaluator

import (
	"fmt"
	"monkey/object"
)

var builtins = map[string]*object.Builtin{
    "puts": {
        Fn: func(args ...object.Object) object.Object {
            for _, obj := range args {
                fmt.Println(obj.Inspect())
            }
            return NULL
        },
    },
    "len": {
        Fn: func(args ...object.Object) object.Object {
            if len(args) != 1 {
                return newError("len() requires exactly one argument, got %d", len(args))
            }

            switch arg := args[0].(type) {
            case *object.Array:
                return &object.Integer{Value: int64(len(arg.Elements))}
            case *object.String:
                return &object.Integer{Value: int64(len(arg.Value))}
            default:
                return newError("argument of type %s not supported for len()", args[0].Type())
            }
        },
    },
    "first": {
        Fn: func(args ...object.Object) object.Object {
            if len(args) != 1 {
                return newError("first() requires exactly one argument, got %d", len(args))
            }

            if args[0].Type() != object.ARRAY_OBJ {
                return newError("argument to first() must be an ARRAY, got %s", args[0].Type())
            }

            arr := args[0].(*object.Array)
            if len(arr.Elements) > 0 {
                return arr.Elements[0]
            }

            return NULL
        },
    },
    "last": {
        Fn: func(args ...object.Object) object.Object {
            if len(args) != 1 {
                return newError("last() requires exactly one argument, got %d", len(args))
            }

            if args[0].Type() != object.ARRAY_OBJ {
                return newError("argument to last() must be an ARRAY, got %s", args[0].Type())
            }

            arr := args[0].(*object.Array)
            length := len(arr.Elements)
            if length > 0 {
                return arr.Elements[length - 1]
            }

            return NULL
        },
    },
    "rest": {
        Fn: func(args ...object.Object) object.Object {
            if len(args) != 1 {
                return newError("rest() requires exactly one argument, got %d", len(args))
            }

            if args[0].Type() != object.ARRAY_OBJ {
                return newError("argument to rest() must be an ARRAY, got %s", args[0].Type())
            }

            arr := args[0].(*object.Array)
            length := len(arr.Elements)
            if length > 0 {
                newElements := make([]object.Object, length-1, length-1)
                copy(newElements, arr.Elements[1:length])
                return &object.Array{Elements: newElements}
            }

            return NULL
        },
    },
    "push": {
        Fn: func(args ...object.Object) object.Object {
            if len(args) != 2 {
                return newError("push() requires exactly two arguments, got %d", len(args))
            }

            if args[0].Type() != object.ARRAY_OBJ {
                return newError("argument to push() must be an ARRAY, got %s", args[0].Type())
            }

            arr := args[0].(*object.Array)
            length := len(arr.Elements)

            newElements := make([]object.Object, length+1, length+1)
            copy(newElements, arr.Elements)
            newElements[length] = args[1]

            return &object.Array{Elements: newElements}
        },
    },
}
