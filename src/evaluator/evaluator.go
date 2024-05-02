package evaluator

import (
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
		return evalStatements(node.Statements)

	case *ast.ExpressionStatement:
		return Eval(node.Expression)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	}

	return nil
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
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
  - Here the method will return a false
  - 2nd call will be true
  - Here the method will return a true

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

func evalStatements(statements []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement)
	}

	return result
}
