package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type ExpectedPrefixTest struct {
	input    string
	expected bool
}

type ExpectedBooleanTest struct {
	input    string
	expected bool
}

type ExpectedFloatTest struct {
	input    string
	expected float64
}

type ExpectedIntegerTest struct {
	input    string
	expected int64
}

type ExpectedIfElseTest struct {
	input    string
	expected interface{}
}

func TestIfElseIfElseExpressions(t *testing.T) {
	tests := []ExpectedIfElseTest{
		{"if (false) { 10 } else if (true) { 11 } else { 12 }", 11},
		{"if (true) { 10 } else if (true) { 11 } else { 12 }", 10},
		{"if (false) { 10 } else if (false) { 11 } else { 12 }", 12},
		{"if (false) { 10 } else if (false) { 11 } else if (true) { 12 } else { 13 }", 12},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)
		integer, ok := tc.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}

}

func TestIfElseExpressions(t *testing.T) {
	tests := []ExpectedIfElseTest{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)
		integer, ok := tc.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}

}

func TestBangOperator(t *testing.T) {
	tests := []ExpectedPrefixTest{
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"!5", false},
		{"!!5", true},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)
		testBooleanObject(t, evaluated, tc.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []ExpectedBooleanTest{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)
		testBooleanObject(t, evaluated, tc.expected)
	}
}

func TestEvalFloatExpression(t *testing.T) {
	tests := []ExpectedFloatTest{
		{"1.5", 1.5},
		{"-1.5", -1.5},
		{"1.5 + 1.5", 3.0},
		{"3.0 - 1.5", 1.5},
		{"1.5 * 3.0", 4.5},
		{"1.5 / 3.0", 0.5},
		{"3.0 % 1.5", 0},
		{"(1.5 + 2) * 4", 14},
		{"1.5 * 4", 6},
		{"4 * 1.5", 6},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)
		testFloatObject(t, evaluated, tc.expected)
	}
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []ExpectedIntegerTest{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"25 % 5", 0},
		{"25 % 4", 1},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"(1 + 2) * 3", 9},
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

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}

	return true
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

func testFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value, got=%0.1f, expected=%0.1f", result.Value, expected)
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
