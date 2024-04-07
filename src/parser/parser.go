package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	lexer *lexer.Lexer

	curToken token.Token
	peekToken token.Token
	errors []string
}

func New(l *lexer.Lexer) *Parser {
	var parser = &Parser{ lexer: l, errors: []string{} }

	// read 2 tokens to curToken and peekToken are initialised
	parser.curToken = parser.lexer.NextToken()
	parser.peekToken = parser.lexer.NextToken()

	return parser
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	var msg = fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)

	p.errors = append(p.errors, msg)
}


func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	var program = &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		var statement = p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}

		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
		case token.LET:
			return p.parseLetStatement()
		default:
			return nil
	}
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}

}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	var statement = &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// We have ran the peek method above, so we are now pointing at the identifier so let x = 5, statement identifier is x
	statement.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

