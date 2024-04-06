package token

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF = "EOF"

	// Identifiers + literals
	IDENT = "IDENT" // add, foobar, x, y - variables
	INT = "INT"

	// Operators
	ASSIGN = "="
	PLUS = "+"
	MINUS = "-"
	BANG = "!"
	ASTERISK = "*"
	SLASH = "/"

	LT = "<"
	GT = ">"
	EQ = "=="
	NOT_EQ = "!="

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
	TRUE = "TRUE"
	FALSE = "FALSE"
	IF = "IF"
	ELSE = "ELSE"
	RETURN = "RETURN"
)

type Token struct {
	Type TokenType
	Literal string
}


var keywords = map[string]TokenType {
	"fn": FUNCTION,
	"let": LET,
	"true": TRUE,
	"false": FALSE,
	"if": IF,
	"else": ELSE,
	"return": RETURN,
}

// Check if the identifier is in the hashmap (fn, let, etc.). If its not in the hashmap, we can assume its a variable name
func LookupIdentifier(identifier string) TokenType {
	// go implicitly returns a boolean to say whether or not the key ok in the map
	// we can do a one-liner, tok will either contain the value or a "zero-value" (the default value for the type, so int will be 0, string will be an empty string)
	// ok (default name) will be either true or false to indicate if the key exists in the map
	if tok, ok := keywords[identifier]; ok {
		return tok
	}

	return IDENT
}