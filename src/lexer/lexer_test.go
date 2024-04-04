package lexer

import (
	"monkey/token"
	"testing"
)

type TokenTest struct {
	expectedType token.TokenType
	expectedLiteral string
}

func TestNextTokenSimple(t *testing.T) {
	var input = "=+(){},;"

	var tests = []TokenTest {
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	var lexedToken = New(input)

	for i, tokenChar := range tests {
		var decodedTokenChar = lexedToken.NextToken()

		if decodedTokenChar.Type != tokenChar.expectedType {
			t.Fatalf("tests[%d] - token_type wrong. expected=%q, got=%q", i, tokenChar.expectedType, decodedTokenChar.Type)
		}

		if decodedTokenChar.Literal != tokenChar.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tokenChar.expectedLiteral, decodedTokenChar.Literal)
		}
	}
}

func TestNextTokenComplex(t *testing.T) {
	var input = `let five = 5;
	let ten = 10;

	let add = fn(x, y) {
		x + y;
	};
	
	let result = add(five, ten);
	`

	var tests = []TokenTest {
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	var lexedToken = New(input)

	for i, tokenChar := range tests {
		var decodedTokenChar = lexedToken.NextToken()

		if decodedTokenChar.Type != tokenChar.expectedType {
			t.Fatalf("tests[%d] - token_type wrong. expected=%q, got=%q", i, tokenChar.expectedType, decodedTokenChar.Type)
		}

		if decodedTokenChar.Literal != tokenChar.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tokenChar.expectedLiteral, decodedTokenChar.Literal)
		}
	}


} 
