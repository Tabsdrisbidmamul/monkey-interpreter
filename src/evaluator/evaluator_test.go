package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type ExpectedIntegerTest struct {
	input    string
	expected int64
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []ExpectedIntegerTest{
		{"5", 5},
		{"10", 10},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)
		testIntegerObject(t, evaluated, tc.expected)
	}
}

func testEval(input string) object.Object {
	lexer := lexer.New(input)
	parser := parser.New(lexer)
	program := parser.ParseProgram()

	return Eval(program)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value, got=%d, expected=%d", result.Value, expected)
		return false
	}

	return true
}
