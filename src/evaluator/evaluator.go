package evaluator

import (
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

// We need to pass the concrete type ast.Node, for all other structs that implements ast.Node to allow the "polymorphism" to work
func Eval(node ast.Node) object.Object {

	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements)

	case *ast.ExpressionStatement:
		return Eval(node.Expression)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.Boolean:
		return nativeToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(node.Operator, left, right)

	case *ast.BlockStatement:
		return evalBlockStatement(node)

	case *ast.IfExpression:
		return evalIfExpression(node)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		return &object.ReturnValue{Value: val}
	}

	return nil
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
func evalIfExpression(ife *ast.IfExpression) object.Object {

	condition := Eval(ife.Condition)

	if isTruthy(condition) {
		return Eval(ife.Consequence)
	} else if len(ife.ElseIfs) > 0 {
		for i := 0; i < len(ife.ElseIfs); i++ {
			eife := ife.ElseIfs[i]
			condition := Eval(eife.Condition)

			if isTruthy(condition) {
				return Eval(eife.Consequence)
			}
		}

		if ife.Alternative != nil {
			return Eval(ife.Alternative)
		} else {
			log.Println("we're in")
			return NULL
		}

	} else if ife.Alternative != nil {
		return Eval(ife.Alternative)
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
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)

	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)

		// left - float right - int
		// left - int right float
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ || left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatIntegerInfixExpression(operator, left, right)

	case operator == "==":
		return nativeToBooleanObject(left == right)
	case operator == "!=":
		return nativeToBooleanObject(left != right)
	default:
		return NULL
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
		return NULL
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
		return NULL
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
		return NULL
	}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperator(right)
	default:
		return NULL
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
		return NULL
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

func evalProgram(statements []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement)

		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement)

		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}
	}

	return result
}
