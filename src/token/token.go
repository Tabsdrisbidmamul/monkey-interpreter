package token

type TokenType string

type Token struct {
	Type TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF = "EOF"

	// Identifiers + literalls
	IDENT = "IDENT" // add, foobar, x, y - variables
	INT = "INT"

	// Operators
	ASSIGN = "="
	PLUS = "+"

	// Delimiters
	COMMA = ","
	SEMICOLON = ";"

	// Special Characters
	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "{"

	// Keywords
	FUNCTION = "FUNCTION"
	LET = "LET"
)