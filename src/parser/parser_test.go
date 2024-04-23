package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"strconv"
	"testing"
)

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

	if !testConsequence(t, *expression, "x") {
		return
	}

	if !testAlternative(t, *expression, "y") {
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

	if !testConsequence(t, *expression, "x") {
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
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
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
	var input = `
let x = 5;
let y = 10;
let foobar = 838383;
	`

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

	var lexer = lexer.New(input)
	var parser = New(lexer)

	var program = parser.ParseProgram()
	checkParserErrors(t, parser)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got%d", len(program.Statements))
	}

	var tests = []ExpectedIdentifierTest{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tok := range tests {
		var statement = program.Statements[i]
		if !testLetStatement(t, statement, tok.expectedIdentifier) {
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

func testAlternative(t *testing.T, expression ast.IfExpression, ident string) bool {
	if len(expression.Alternative.Statements) != 1 {
		t.Errorf("alternative is not 1 statement. got=%d\n", len(expression.Alternative.Statements))
		return false
	}

	alternative, ok := expression.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expression.Alternative.Statements[0] is not ast.ExpressionStatement. got=%T", expression.Alternative.Statements[0])
		return false
	}

	if !testIdentifier(t, alternative.Expression, "y") {
		return false
	}

	return true
}

func testConsequence(t *testing.T, expression ast.IfExpression, ident string) bool {
	if len(expression.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statement. got=%d\n", len(expression.Consequence.Statements))
	}

	consequence, ok := expression.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expression.Consequence.Statements[0] is not ast.ExpressionStatement. got=%T", expression.Consequence.Statements[0])
		return false
	}

	if !testIdentifier(t, consequence.Expression, ident) {
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
