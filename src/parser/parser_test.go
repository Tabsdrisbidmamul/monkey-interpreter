package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"strconv"
	"testing"
)

type ExpectedLetStatementTest struct {
	input              string
	expectedIdentifier string
	expectedValue      interface{}
}

type ExpectedCallArgumentTest struct {
	input         string
	expectedIdent string
	expectedArgs  []string
}

type ExpectedParameterTest struct {
	input          string
	expectedParams []string
}

type ExpectedPrecedenceTest struct {
	input    string
	expected string
}

type ExpectedBooleanTest struct {
	input           string
	expectedBoolean bool
}

type ExpectedIdentifierTest struct {
	expectedIdentifier string
}

type ExpectedPrefixTest struct {
	input    string
	operator string
	value    interface{}
}

type ExpectedInfixTest struct {
	input      string
	leftValue  interface{}
	operator   string
	rightValue interface{}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := statement.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expression not *ast.StringLiteral. got=%T", statement.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestCallParameterParsing(t *testing.T) {
	tests := []ExpectedCallArgumentTest{
		{
			input:         "add();",
			expectedIdent: "add",
			expectedArgs:  []string{},
		},
		{
			input:         "add(1);",
			expectedIdent: "add",
			expectedArgs:  []string{"1"},
		},
		{
			input:         "add(1, 2 * 3, 4 + 5);",
			expectedIdent: "add",
			expectedArgs:  []string{"1", "(2 * 3)", "(4 + 5)"},
		},
	}

	for _, tc := range tests {
		lexer := lexer.New(tc.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)

		statement := program.Statements[0].(*ast.ExpressionStatement)
		exp, ok := statement.Expression.(*ast.CallExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
				statement.Expression)
		}

		if !testIdentifier(t, exp.Function, tc.expectedIdent) {
			return
		}

		if len(exp.Arguments) != len(tc.expectedArgs) {
			t.Fatalf("wrong number of arguments. want=%d, got=%d",
				len(tc.expectedArgs), len(exp.Arguments))
		}

		for i, arg := range tc.expectedArgs {
			if exp.Arguments[i].String() != arg {
				t.Errorf("argument %d wrong. want=%q, got=%q", i,
					arg, exp.Arguments[i].String())
			}
		}
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5)"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements expected %d statements. got=%d", 1, len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	exp, ok := statement.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("statement.Expression is not CallExpression. got=%T", statement.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("expected %d arguments. got=%d", 3, len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)

}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []ExpectedParameterTest{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tc := range tests {
		lexer := lexer.New(tc.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)

		statement := program.Statements[0].(*ast.ExpressionStatement)
		function := statement.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(tc.expectedParams) {
			t.Errorf("length parameters wrong, expected %d, got=%d", len(tc.expectedParams), len(function.Parameters))
		}

		for i, ident := range tc.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	function, ok := statement.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("statement.Expression is not ast.FunctionLiteral. got=%T", statement.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. expected %d, got=%d\n", 2, len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements expected %d statement. got=%d\n", 1, len(function.Body.Statements))
	}

	bodyStatement, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body statement is not ast.ExpressionStatement. got=%T", function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStatement.Expression, "x", "+", "y")

}

func TestIfElifElseExpression(t *testing.T) {
	input := `if (x < y) { x } else if (x == y) { y } else { z }`

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	ifStatement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ifExpression, ok := ifStatement.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("statement.Expression is not ast.IfExpression. got=%T", ifStatement.Expression)
	}

	if len(ifExpression.ElseIfs) != 1 {
		t.Fatalf("ifExpression.ElseIfs is not %d. got=%d", 1, len(ifExpression.ElseIfs))
	}

	if !testInfixExpression(t, ifExpression.Condition, "x", "<", "y") {
		return
	}

	if !testConsequence(t, *ifExpression.Consequence, "x") {
		return
	}

	if !testInfixExpression(t, ifExpression.ElseIfs[0].Condition, "x", "==", "y") {
		return
	}

	if !testConsequence(t, *ifExpression.ElseIfs[0].Consequence, "y") {
		return
	}

	if !testAlternative(t, *ifExpression.Alternative, "z") {
		return
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	expression, ok := statement.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("statement.Expression is not ast.IfExpression. got=%T", statement.Expression)
	}

	if !testInfixExpression(t, expression.Condition, "x", "<", "y") {
		return
	}

	if !testConsequence(t, *expression.Consequence, "x") {
		return
	}

	if !testAlternative(t, *expression.Alternative, "y") {
		return
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	expression, ok := statement.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("statement.Expression is not ast.IfExpression. got=%T", statement.Expression)
	}

	if !testInfixExpression(t, expression.Condition, "x", "<", "y") {
		return
	}

	if !testConsequence(t, *expression.Consequence, "x") {
		return
	}

	if expression.Alternative != nil {
		t.Errorf("expression.Alternative.Statements was not nil. got=%+v", expression.Alternative)
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []ExpectedBooleanTest{
		{input: "true;", expectedBoolean: true},
		{input: "false;", expectedBoolean: false},
	}

	for _, tc := range tests {
		lexer := lexer.New(tc.input)
		parser := New(lexer)
		program := parser.ParseProgram()

		checkParserErrors(t, parser)

		if len(program.Statements) != 1 {
			t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		testLiteralExpression(t, statement.Expression, tc.expectedBoolean)
	}

}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []ExpectedPrecedenceTest{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a % b * c",
			"((a % b) * c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"-1 * 2 + 3",
			"(((-1) * 2) + 3)",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"(1.5 + 1.5) * 2",
			"((1.5 + 1.5) * 2)",
		},
	}

	for _, tc := range tests {
		lexer := lexer.New(tc.input)
		parser := New(lexer)
		program := parser.ParseProgram()
		checkParserErrors(t, parser)

		actual := program.String()
		if actual != tc.expected {
			t.Errorf("expected=%q, got=%q", tc.expected, actual)
		}
	}
}

// Explanation in parseExpression
func TestParsingInfixExpression(t *testing.T) {
	infixTests := []ExpectedInfixTest{
		{"5 + 5;", 5, "+", 5},
		{"1.5 + 1.5;", 1.5, "+", 1.5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 % 5;", 5, "%", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, tc := range infixTests {
		lexer := lexer.New(tc.input)
		parser := New(lexer)
		program := parser.ParseProgram()

		checkParserErrors(t, parser)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		if !testInfixExpression(t, statement.Expression, tc.leftValue, tc.operator, tc.rightValue) {
			return
		}
	}
}

func TestParsingPrefixExpression(t *testing.T) {
	prefixTests := []ExpectedPrefixTest{
		{"!5", "!", 5},
		{"-15", "-", 15},
		{"-1.5", "-", 1.5},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, testCase := range prefixTests {
		lexer := lexer.New(testCase.input)
		parser := New(lexer)
		program := parser.ParseProgram()

		checkParserErrors(t, parser)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		if !testPrefixExpression(t, statement.Expression, testCase.operator, testCase.value) {
			return
		}
	}
}

func TestFloatLiteralExpression(t *testing.T) {
	input := "1.2"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	if !testLiteralExpression(t, statement.Expression, 1.2) {
		return
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	if !testLiteralExpression(t, statement.Expression, 5) {
		return
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar"

	// In the program.Statements[] we have got one ExpressionStatement, the tree below is what we have got
	/* Statements
				 |
	ExpressionStatement
				 |
	Token  Expression (Identifier implements Expression)
						|
					Identifier
					   |
					Token Value
	*/

	lexer := lexer.New(input)
	parser := New(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	if !testLiteralExpression(t, statement.Expression, input) {
		return
	}
}

func TestReturnStatement(t *testing.T) {
	input := `
return 5;
return 10;
return 993322;
	`

	lexer := lexer.New(input)
	parser := New(lexer)

	program := parser.ParseProgram()
	checkParserErrors(t, parser)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	for _, statement := range program.Statements {
		returnStatement, ok := statement.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("returnStatement not *ast.ReturnStatement. got=%T", returnStatement)
			continue
		}

		if returnStatement.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return', got %q",
				returnStatement.TokenLiteral())
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []ExpectedLetStatementTest{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	// let x = 5
	// The node looks likes
	/* []Statements {
		&LetStatement {
			Token: {},
			Name: Value: &Identifier {
				Token: {},
				Value: {}
			},
			Value: &Identifier {
				Token: {},
				Value: {}
			}
		}
	}

	There can n number LetStatements
	statements
			|
	LetStatements
		|		  |     |
	Token  Name Value
								|
						Token Value
	*/
	// Token = token.Token {Type: token.LET, Literal: "let" }
	// Name = &Identifier { Type: token.IDENT, Literal: "x" }
	// Value = &Identifier { Type: token.INT, Literal: "5" }

	for _, tc := range tests {
		var lexer = lexer.New(tc.input)
		var parser = New(lexer)

		var program = parser.ParseProgram()
		checkParserErrors(t, parser)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}

		statement := program.Statements[0]
		if !testLetStatement(t, statement, tc.expectedIdentifier) {
			return
		}

		value := statement.(*ast.LetStatement).Value
		if !testLiteralExpression(t, value, tc.expectedValue) {
			return
		}
	}

}

func TestBadParseAndSucceed(t *testing.T) {
	var input = `
let x 5;
let = 10;
let 838383;
		`

	var lexer = lexer.New(input)
	var parser = New(lexer)

	var program = parser.ParseProgram()

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(parser.Errors()) < 3 {
		t.Errorf("parser.error not 3, got=%d", len(parser.Errors()))
	}
}

// ------Helpers---------
func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case float64:
		return testFloatLiteral(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	}

	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testPrefixExpression(t *testing.T, exp ast.Expression, operator string, right interface{}) bool {
	expression, ok := exp.(*ast.PrefixExpression)
	if !ok {
		t.Fatalf("exp is not ast.PrefixExpression. got=%T", exp)
		return false
	}

	if expression.Operator != operator {
		t.Fatalf("expression.Operator is not %s. got=%s", operator, expression.Operator)
		return false
	}

	if !testLiteralExpression(t, expression.Right, right) {
		return false
	}

	return true
}

func testAlternative(t *testing.T, alternative ast.BlockStatement, ident string) bool {
	if len(alternative.Statements) != 1 {
		t.Errorf("alternative is not 1 statement. got=%d\n", len(alternative.Statements))
		return false
	}

	alternativeExpression, ok := alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expression.Alternative.Statements[0] is not ast.ExpressionStatement. got=%T", alternative.Statements[0])
		return false
	}

	if !testIdentifier(t, alternativeExpression.Expression, ident) {
		return false
	}

	return true
}

func testConsequence(t *testing.T, consequence ast.BlockStatement, ident string) bool {
	if len(consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statement. got=%d\n", len(consequence.Statements))
	}

	expression, ok := consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expression.Consequence.Statements[0] is not ast.ExpressionStatement. got=%T",
			consequence.Statements[0])
		return false
	}

	if !testIdentifier(t, expression.Expression, ident) {
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	boolean, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if boolean.Value != value {
		t.Errorf("boolean.Value not %t. got=%t", value, boolean.Value)
		return false
	}

	if boolean.TokenLiteral() != strconv.FormatBool(value) {
		t.Errorf("boolean.TokenLiteral() not %t. got=%s", value, boolean.TokenLiteral())
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	identifier, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if identifier.Value != value {
		t.Errorf("identifier.Value not %s. got=%s", value, identifier.Value)
		return false
	}

	if identifier.TokenLiteral() != value {
		t.Errorf("identifier.TokenLiteral() not %s. got=%s", value, identifier.TokenLiteral())
		return false
	}

	return true
}

func testFloatLiteral(t *testing.T, floatLiteral ast.Expression, value float64) bool {
	_float, ok := floatLiteral.(*ast.FloatLiteral)
	if !ok {
		t.Errorf("floatLiteral not *ast.FloatLiteral. got=%T", floatLiteral)
		return false
	}

	if _float.Value != value {
		t.Errorf("_float.Value not %f, got=%f", value, _float.Value)
		return false
	}

	if _float.TokenLiteral() != fmt.Sprintf("%0.1f", value) {
		t.Errorf("_float.TokenLiteral() not %0.1f, got=%s", value, _float.TokenLiteral())
		return false
	}

	return true
}

func testIntegerLiteral(t *testing.T, integerLiteral ast.Expression, value int64) bool {
	_integer, ok := integerLiteral.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("integerLiteral not *ast.IntegerLiteral. got=%T", integerLiteral)
		return false
	}

	if _integer.Value != value {
		t.Errorf("_integer.Value not %d. got=%d", value, _integer.Value)
		return false
	}

	if _integer.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("_integer.TokenLiteral() not %d. got=%s", value, _integer.TokenLiteral())
		return false
	}

	return true
}

func testLetStatement(t *testing.T, statement ast.Statement, expectedIdentifier string) bool {
	if statement.TokenLiteral() != "let" {
		t.Errorf("statement.TokenLiteral() not 'let'. got=%q", statement.TokenLiteral())
		return false
	}

	letStatement, ok := statement.(*ast.LetStatement)
	if !ok {
		t.Errorf("statement not *ast.LetStatement. got=%T", statement)
		return false
	}

	if letStatement.Name.Value != expectedIdentifier {
		t.Errorf("letStatement.Name.Value not '%s'. got=%s", expectedIdentifier, letStatement.Name.Value)
		return false
	}

	if letStatement.Name.TokenLiteral() != expectedIdentifier {
		t.Errorf("letStatement.Name.TokenLiteral() not '%s'. got=%s", expectedIdentifier, letStatement.Name.TokenLiteral())
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	var errors = p.Errors()

	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))

	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}

	t.FailNow()
}
