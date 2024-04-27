package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type ExpectedBooleanTest struct {
	input    string
	expected bool
}

type ExpectedIntegerTest struct {
	input    string
	expected int64
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []ExpectedBooleanTest{

		{"false", false},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)
		testBooleanObject(t, evaluated, tc.expected)
	}
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

// --------HELPERS------------------
func testEval(input string) object.Object {
	lexer := lexer.New(input)
	parser := parser.New(lexer)
	program := parser.ParseProgram()

	return Eval(program)
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)

	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value, got=%t, expected=%t", result.Value, expected)
		return false
	}

	return true
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
