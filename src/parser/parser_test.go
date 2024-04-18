package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

type ExpectedIdentifierTest struct {
	expectedIdentifier string
}

type ExpectedPrefixTest struct {
	input        string
	operator     string
	integerValue int64
}

type ExpectedInfixTest struct {
	input      string
	leftValue  int64
	operator   string
	rightValue int64
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
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

		expression, ok := statement.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("expression is not ast.InfixExpression. got=%T", statement.Expression)
		}

		// assert left value
		if !testIntegerLiteral(t, expression.Left, tc.leftValue) {
			return
		}

		if expression.Operator != tc.operator {
			t.Fatalf("expression.Operator is not %s. got=%s", tc.operator, expression.Operator)
		}

		// assert right value
		if !testIntegerLiteral(t, expression.Right, tc.rightValue) {
			return
		}
	}
}

func TestParsingPrefixExpression(t *testing.T) {
	prefixTests := []ExpectedPrefixTest{
		{"!5", "!", 5},
		{"-15", "-", 15},
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

		expression, ok := statement.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("statement is not ast.PrefixExpression. got=%T", statement.Expression)
		}

		if expression.Operator != testCase.operator {
			t.Fatalf("expression.Operator is not %s. got=%s", testCase.operator, expression.Operator)
		}

		if !testIntegerLiteral(t, expression.Right, testCase.integerValue) {
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

	literal, ok := statement.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expression not *ast.IntegerLiteral. got=%T", program.Statements[0])
	}

	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}

	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "5",
			literal.TokenLiteral())
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

	identifier, ok := statement.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", statement.Expression)
	}

	if identifier.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", identifier.Value)
	}

	if identifier.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "foobar",
			identifier.TokenLiteral())
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
