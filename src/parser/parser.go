package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	// setup "ENUM" values to start from int 0, so LOWEST is int 1
	// This retains the order, which will be used for precedence ordering
	// larger the value, the higher the precedence it has
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > OR <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunc(X)
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	lexer *lexer.Lexer

	curToken  token.Token
	peekToken token.Token
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	var parser = &Parser{lexer: l, errors: []string{}}

	// read 2 tokens to curToken and peekToken are initialised
	parser.curToken = parser.lexer.NextToken()
	parser.peekToken = parser.lexer.NextToken()

	// initialise the prefixParseFns map, and register all prefixes to map
	parser.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	parser.registerPrefix(token.IDENT, parser.parseIdentifier)
	parser.registerPrefix(token.INT, parser.parseIntegerLiteral)
	parser.registerPrefix(token.BANG, parser.parsePrefixExpression)
	parser.registerPrefix(token.MINUS, parser.parsePrefixExpression)
	parser.registerPrefix(token.TRUE, parser.parseBoolean)
	parser.registerPrefix(token.FALSE, parser.parseBoolean)

	// initialise the infixParseFns map, and register all infixes to maps
	parser.infixParseFns = make(map[token.TokenType]infixParseFn)
	parser.registerInfix(token.PLUS, parser.parseInfixExpression)
	parser.registerInfix(token.MINUS, parser.parseInfixExpression)
	parser.registerInfix(token.SLASH, parser.parseInfixExpression)
	parser.registerInfix(token.ASTERISK, parser.parseInfixExpression)
	parser.registerInfix(token.EQ, parser.parseInfixExpression)
	parser.registerInfix(token.NOT_EQ, parser.parseInfixExpression)
	parser.registerInfix(token.LT, parser.parseInfixExpression)
	parser.registerInfix(token.GT, parser.parseInfixExpression)

	return parser
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
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: p.curToken}

	statement.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
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

// The Identifier type implements the Expression interface
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// fmt.Printf("curToken is %+v\n", p.curToken)

	/*
		      -5
		      PrefixExpression { Token: {Type: -, Literal: -}, Operator: -, Right: 5 }

		      5 + 5
		      InfixExpression {Token:{Type:+ Literal:+} Left:5 Operator:+ Right:5}

		      The flow for PrefixExpression:
		      - We will get the prefix function, and obtain the left expression struct, the next line will either be a semi colon or the next token's precedence is LOWEST, which will stop lookup

		      - Returning back the PrefixExpression

		      The flow for InfixExpression:
		      - We will get the prefix function, and obtain the left expression struct, but the next line will not contain a semi colon, and the next token's precedence (being a arithmetic operator) will be greater than LOWEST.

		      - We successfully obtain the infix function, move the pointer ahead to the next operand in the calculation.

		      - We pass in the previous left exp into the infix function, and get back the InfixExpression struct

		                                          Left
		      1 + 2 + 3 will be serialised to ((1 + 2) + 3)
		                                                Right
		      AST tree will look

		      InfixExpression struct {
		        Token    token.Token
		        Left     Expression
		        Operator string
		        Right    Expression
		      }

		      InfixExpression
		              |
		      |------------------------ |
		InfixExpression (left)    IntegerLiteral (Right)
		      |                         |
		-------------------             3
		|                 |
		IntegerLiteral   IntegerLiteral
		|                   |
		1                   2


	*/
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()
	// fmt.Printf("leftExp is %+v\n", leftExp)

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// fmt.Printf("p.peekToken is %+v\n", p.peekToken)
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		// move the to the infix operator
		p.nextToken()
		// e.g. 5 + 5, the leftExp will be 5, and curToken is now +
		leftExp = infix(leftExp)
	}

	return leftExp
}

// -----pre/in fix fns--------
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	/*
	   e.g. 5 + 5
	   expression := {Token:{Type:+ Literal:+} Left:5 Operator:+ Right:<nil>}
	*/
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	// fmt.Printf("expression is %+v\n", *expression)

	// get precedence order
	// so 5 + 5, the plus is 4 in the constant enum which is SUM
	precedence := p.curPrecedence()
	// fmt.Printf("precedence is %+v\n", precedence)

	// e.g. 5 + 5, curToken is 5
	p.nextToken()

	expression.Right = p.parseExpression(precedence)
	// fmt.Printf("expression is %+v\n", *expression)

	return expression
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	literal := &ast.Boolean{Token: p.curToken}

	value, err := strconv.ParseBool(p.curToken.Literal)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as boolean", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	literal.Value = value
	return literal
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	literal.Value = value
	return literal
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// ----------Helpers-----------
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) peekError(t token.TokenType) {
	var msg = fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)

	p.errors = append(p.errors, msg)
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

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}
