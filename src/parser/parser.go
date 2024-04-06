package parser

import (
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	lexer *lexer.Lexer

	curToken token.Token
	peekToken token.Token
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	return nil
}

func New(l *lexer.Lexer) *Parser {
	var parser = &Parser{lexer: l}

	// read 2 tokens to curToken and peekToken are initialised
	parser.curToken = parser.lexer.NextToken()
	parser.peekToken = parser.lexer.NextToken()

	return parser
}