package evaluator

import (
	"fmt"
	"log"
	"math"
	"monkey/ast"
	"monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, expected=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}

			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}

			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments.\nexpected=1, got=%d", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to \"first\" must be an ARRAY type.\ngot %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		},
	},
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments.\nexpected=1, got=%d", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to \"last\" must be an ARRAY type.\ngot %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[len(arr.Elements)-1]
			}

			return NULL
		},
	},
}

// We need to pass the concrete type ast.Node, for all other structs that implements ast.Node to allow the "polymorphism" to work
func Eval(node ast.Node, env *object.Environment) object.Object {

	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.Boolean:
		return nativeToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}

		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}

		env.Set(node.Name.Value, val)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpression(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.ArrayLiteral:
		elements := evalExpression(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)

	}

	return nil
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)

	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalExpression(expression []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range expression {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

/*
when the condition is true, we return the result "true"
- if (5 < 10) {return "true"} else { return "false" }

# In our code, we will evaluate the consequence

when the condition is false, re return result "false"
- if (5 > 10) {return "true"} else { return "false" }

# In our code, we will evaluate the alternative

when the condition is false, but no else block is given, we return false
if (5 < 10) {return "true"}

In our code, we will return NULL
*/
func evalIfExpression(ife *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ife.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ife.Consequence, env)
	} else if len(ife.ElseIfs) > 0 {
		for i := 0; i < len(ife.ElseIfs); i++ {
			eife := ife.ElseIfs[i]
			condition := Eval(eife.Condition, env)

			if isTruthy(condition) {
				return Eval(eife.Consequence, env)
			}
		}

		if ife.Alternative != nil {
			return Eval(ife.Alternative, env)
		} else {
			log.Println("we're in")
			return NULL
		}

	} else if ife.Alternative != nil {
		return Eval(ife.Alternative, env)
	} else {
		return NULL
	}
}

func isTruthy(condition object.Object) bool {
	switch condition {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}

}

func evalInfixExpression(operator string, left, right object.Object) object.Object {

	switch {
	// integer
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)

		// float
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)

		// left - float right - int
		// left - int right float
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ || left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatIntegerInfixExpression(operator, left, right)

		// string evaluations
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:

		return evalStringInfixExpression(operator, left, right)

	case operator == "==":
		return nativeToBooleanObject(left == right)
	case operator == "!=":
		return nativeToBooleanObject(left != right)

	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())

	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "==":
		return &object.Boolean{Value: leftVal == rightVal}

	case "+":
		return &object.String{Value: leftVal + rightVal}

	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

}

func evalFloatIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	var _integerValue int64
	var _floatValue float64

	_, leftFloatErr := left.(*object.Float)
	if !leftFloatErr {
		_integerValue = left.(*object.Integer).Value
		_floatValue = right.(*object.Float).Value
	}

	_, rightFloatErr := right.(*object.Float)
	if !rightFloatErr {
		_floatValue = left.(*object.Float).Value
		_integerValue = right.(*object.Integer).Value
	}

	switch operator {
	case "+":
		return &object.Float{Value: _floatValue + float64(_integerValue)}
	case "-":
		return &object.Float{Value: _floatValue - float64(_integerValue)}
	case "*":
		return &object.Float{Value: _floatValue * float64(_integerValue)}
	case "/":
		return &object.Float{Value: _floatValue / float64(_integerValue)}
	case "%":
		return &object.Float{Value: math.Mod(_floatValue, float64(_integerValue))}
	case "<":
		return nativeToBooleanObject(_floatValue < float64(_integerValue))
	case ">":
		return nativeToBooleanObject(_floatValue > float64(_integerValue))
	case "==":
		return nativeToBooleanObject(_floatValue == float64(_integerValue))
	case "!=":
		return nativeToBooleanObject(_floatValue != float64(_integerValue))
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "%":
		return &object.Float{Value: math.Mod(leftVal, rightVal)}
	case "<":
		return nativeToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "%":
		return &object.Integer{Value: leftVal % rightVal}
	case "<":
		return nativeToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperator(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalMinusPrefixOperator(right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER_OBJ:
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	case object.FLOAT_OBJ:
		value := right.(*object.Float).Value
		return &object.Float{Value: -value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

/*
The way this works
!true == false
the right value is false, so the switch cases will go to true and return false (simple inversion we manually coded)

same goes for !false == true

# Any other object (node) will be defaulted to false

!!true == true
the eval step in case *ast.PrefixExpression in Eval() method, it wil recursively call Eval on the right node, in this case its (!(!true)), so the
  - 1st right is !true
    _ Here the method will return a false
  - 2nd call will be true
    _ Which will go to the Boolean case switch

The first return is passed down (which was false), in our lookup table it wil return true.

The same story for !!false == false
*/
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func nativeToBooleanObject(value bool) *object.Boolean {
	if value {
		return TRUE
	}

	return FALSE
}

func evalProgram(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}

	}

	return result
}

// error helpers
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
