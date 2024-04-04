package lexer

import "monkey/token"

/*
	position is the current point and readPosition is position + 1
	we need to peek further in the input
*/
type Lexer struct {
	input string
	position int // current position in input - points to current char
	readPosition int // current reading position in input - after current char
	ch byte // current char under examination
}

func New(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.readChar()
	return lexer
}

func (lexer *Lexer) readChar() {
	if lexer.readPosition >= len(lexer.input) {
		lexer.ch = 0 // ASCII code for NUL - so EOF or nothing read in
	} else {
		lexer.ch = lexer.input[lexer.readPosition]
	}

	lexer.position = lexer.readPosition
	lexer.readPosition += 1
}

func (l *Lexer) NextToken() token.Token { 
	var _token token.Token

	switch l.ch {
		case '=':
			_token = newToken(token.ASSIGN, l.ch)
		case ';':
			_token = newToken(token.SEMICOLON, l.ch)
		case '(':
			_token = newToken(token.LPAREN, l.ch)
		case ')':
			_token = newToken(token.RPAREN, l.ch)
		case ',':
			_token = newToken(token.COMMA, l.ch)
		case '+':
			_token = newToken(token.PLUS, l.ch)
		case '{':
			_token = newToken(token.LBRACE, l.ch)
		case '}':
			_token = newToken(token.RBRACE, l.ch)
		case 0:
			_token.Literal = ""
			_token.Type = token.EOF
		default:
			if isLetter(l.ch) {
				_token.Literal = l.readIdentifier()
				return _token
			}

			_token = newToken(token.ILLEGAL, l.ch)
		}

		l.readChar()
		return _token
}

// Read each character in the variable identifier moving along and return the ending position of the identifier 
func (l *Lexer) readIdentifier() string {
	var position = l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position: l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

