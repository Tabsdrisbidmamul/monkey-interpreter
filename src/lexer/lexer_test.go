package lexer

import (
	"monkey/token"
	"testing"
)

type TokenTest struct {
	expectedType    token.TokenType
	expectedLiteral string
}

func TestIfElifElseTokens(t *testing.T) {
	input := `
  if (5 < 10) {
    return true;
  } else if (6 < 10) {
    return true;
  } else {
    return false;
  }
  `

	tests := []TokenTest{
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},

		{token.ELSEIF, "else if"},
		{token.LPAREN, "("},
		{token.INT, "6"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},

		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
	}

	lexedToken := New(input)

	testLexedToken(t, lexedToken, tests)
}

func TestReadInFloatAndInteger(t *testing.T) {
	input := `let fl = 1.234;
  let integer = 1234;
  `

	tests := []TokenTest{
		{token.LET, "let"},
		{token.IDENT, "fl"},
		{token.ASSIGN, "="},
		{token.FLOAT, "1.234"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "integer"},
		{token.ASSIGN, "="},
		{token.INT, "1234"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	lexedToken := New(input)

	testLexedToken(t, lexedToken, tests)
}

func TestStringToken(t *testing.T) {
	input := `"foobar";
  `

	tests := []TokenTest{
		{token.STRING, "foobar"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	lexedToken := New(input)

	testLexedToken(t, lexedToken, tests)
}

func TestNextTokenSimple(t *testing.T) {
	var input = "=+(){},*/%;"

	var tests = []TokenTest{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.MOD, "%"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	var lexedToken = New(input)

	testLexedToken(t, lexedToken, tests)
}

func TestNextTokenComplex(t *testing.T) {
	var input = `let five = 5;
let ten = 10;

let add = fn(x, y) {
	x + y;
};

let result = add(five, ten);
!-/*5;
5 < 10 > 5;

if (5 < 10) {
return true;
} else {
return false;
}

10 == 10;
10 != 9;
"foobar"
"foo bar"
[1, 2];
{"foo": "bar"}
`

	var tests = []TokenTest{
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

		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},

		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},

		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},

		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},

		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},

		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},

		{token.EOF, ""},
	}

	var lexedToken = New(input)

	testLexedToken(t, lexedToken, tests)
}

func testLexedToken(t *testing.T, l *Lexer, tests []TokenTest) {
	for i, tokenChar := range tests {
		var decodedTokenChar = l.NextToken()

		if decodedTokenChar.Type != tokenChar.expectedType {
			t.Fatalf("tests[%d] - token_type wrong. expected=%q, got=%q.\nFor token %+v", i, tokenChar.expectedType, decodedTokenChar.Type, tokenChar)
		}

		if decodedTokenChar.Literal != tokenChar.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tokenChar.expectedLiteral, decodedTokenChar.Literal)
		}
	}
}
