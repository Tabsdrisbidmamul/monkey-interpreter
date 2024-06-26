package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type ExpectedTest[T any] struct {
	input    string
	expected T
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []ExpectedTest[interface{}]{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
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

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. \nexpected=3\ngot=%d", len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []ExpectedTest[interface{}]{
		// string len
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, expected=1"},

		// array len
		{`len([1, 2, 3])`, 3},
		{`len([1, 2])`, 2},
		{`len([1])`, 1},
		{`len([])`, 0},

		// array first
		{`first([1, 2, 3])`, 1},
		{`first([1])`, 1},
		{`first()`, "wrong number of arguments.\nexpected=1, got=0"},
		{`first(5)`, "argument to \"first\" must be an ARRAY type.\ngot INTEGER"},

		// array last
		{`last([1, 2, 3])`, 3},
		{`last([1])`, 1},
		{`last()`, "wrong number of arguments.\nexpected=1, got=0"},
		{`last(5)`, "argument to \"last\" must be an ARRAY type.\ngot INTEGER"},

		// array rest
		{
			`rest([1, 2, 3])`,
			&object.Array{
				Elements: []object.Object{
					&object.Integer{Value: 2},
					&object.Integer{Value: 3},
				},
			},
		},
		{
			`rest([1])`,
			&object.Array{
				Elements: []object.Object{},
			},
		},
		{`rest()`, "wrong number of arguments.\nexpected=1, got=0"},
		{`rest(5)`, "argument to \"rest\" must be an ARRAY type.\ngot INTEGER"},

		// array push
		{
			`push([], 2)`,
			[]int64{2},
		},
		{
			`push([1], 2)`,
			[]int64{1, 2},
		},
		{
			`push([])`,
			"wrong number of arguments.\nexpected=2, got=1",
		},
		{
			`push(5, 5)`,
			"argument to \"push\" must be an ARRAY type.\ngot INTEGER",
		},

		// array pop
		{
			`pop([1, 2])`,
			2,
		},
		{
			`pop([1])`,
			1,
		},
		{
			`pop([])`,
			"cannot pop an empty array",
		},
		{
			`pop()`,
			"wrong number of arguments.\nexpected=1, got=0",
		},
		{
			`pop(5)`,
			"argument to \"pop\" must be an ARRAY type.\ngot INTEGER",
		},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)

		switch expected := tc.expected.(type) {
		case []int64:
			testArrayObject(t, evaluated, expected)
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		}
	}
}

func TestStringComparison(t *testing.T) {
	input := `"test" == "test"`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Boolean)
	if !ok {
		t.Fatalf("object is not Boolean. got=%T (%+v)", evaluated, evaluated)
	}

	if result.Value != true {
		t.Errorf("Boolean has wrong value. expected=%t, got=%t", true, result.Value)
	}

}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. expected=%q, got=%q", "Hello World!", str.Value)
	}

}
func TestClosures(t *testing.T) {
	input := `
let newAdder = fn(x) {
  fn(y) { x + y };
};
let addTwo = newAdder(2);
addTwo(2);`

	testIntegerObject(t, testEval(input), 4)
}

func TestFunctionAppLiteral(t *testing.T) {
	tests := []ExpectedTest[int64]{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tc := range tests {
		testIntegerObject(t, testEval(tc.input), tc.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}

}

func TestLetStatements(t *testing.T) {
	tests := []ExpectedTest[int64]{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tc := range tests {
		testIntegerObject(t, testEval(tc.input), tc.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []ExpectedTest[string]{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
    if (10 > 1) {
      if (10 > 1) {
        return true + false;
      }
        return 1;
    }
      `,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if errObj.Message != tc.expected {
			t.Errorf("wrong error message. expected=%s, got=%s", tc.expected, errObj.Message)

		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []ExpectedTest[int64]{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`
    if (10 > 1) {
      if (10 > 1) {
        return 10;
      }

      return 1;
    }
    `, 10},
		{
			`
let f = fn(x) {
  return x;
  x + 10;
};
f(10);`,
			10,
		},
		{
			`
let f = fn(x) {
   let result = x + 10;
   return result;
   return 10;
};
f(10);`,
			20,
		},
	}

	for _, tc := range tests {
		evaluated := testEval(tc.input)
		testIntegerObject(t, evaluated, tc.expected)
	}
}

func TestIfElseIfElseExpressions(t *testing.T) {
	tests := []ExpectedTest[interface{}]{
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
	tests := []ExpectedTest[interface{}]{
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
	tests := []ExpectedTest[bool]{
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
	tests := []ExpectedTest[bool]{
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
	tests := []ExpectedTest[float64]{
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
	tests := []ExpectedTest[int64]{
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
	env := object.NewEnvironment()

	return Eval(program, env)
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

func testArrayObject(t *testing.T, obj object.Object, expected []int64) bool {
	result, ok := obj.(*object.Array)
	if !ok {
		t.Errorf("object is not Array. got=%T (%+v)", obj, obj)
		return false
	}

	for index := range result.Elements {
		expectedValue := expected[index]

		actualValue := result.Elements[index].(*object.Integer).Value

		if actualValue != expectedValue {
			t.Errorf("actual Array.Elements has the wrong values\nexpected=%d at index=%d.\ngot=%d at index=%d", expectedValue, index, actualValue, index)
		}
	}

	return true
}
