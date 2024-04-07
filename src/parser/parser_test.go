package parser

import (
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

type ExpectedIdentifierTest struct {
	expectedIdentifier string
}

func TestLetStatements(t *testing.T) {
	var input = `
let x = 5;
let y = 10;
let foobar = 838383;
	`

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

	var tests = []ExpectedIdentifierTest {
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