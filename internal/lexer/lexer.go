// Package lexer tokeniza texto SQL en internal/lexer -> internal/parser.
//
// PROPIEDAD: Maicol.
// Esqueleto de partida: define el tipo Token y sus tipos, sin implementar
// aún el escaneo real. Ver docs/gramatica.md para el subconjunto de SQL a
// cubrir.

// Package lexer tokeniza texto SQL en internal/lexer -> internal/parser.
//
// Package lexer tokeniza texto SQL en internal/lexer -> internal/parser.
//
// PROPIEDAD: Maicol.
package lexer

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenInt
	TokenFloat
	TokenString
	TokenKeyword
	TokenOperator
	TokenIllegal
)

type Token struct {
	Type TokenType
	Lit  string
	Pos  int
	Line int
	Col  int
}

var keywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true,
	"AND": true, "OR": true, "TRUE": true, "FALSE": true, "NULL": true,
	"ORDER": true, "BY": true, "LIMIT": true, "ASC": true, "DESC": true,
	"GROUP": true, "COUNT": true, "SUM": true, "AVG": true, "MIN": true, "MAX": true,
	"INNER": true, "JOIN": true, "ON": true,
}

type Lexer struct {
	input        string
	position     int  // posición actual
	readPosition int  // siguiente posición a leer
	ch           byte // carácter actual
	line         int
	col          int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.col++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	tok.Line = l.line
	tok.Col = l.col
	tok.Pos = l.position

	switch l.ch {
	case '*':
		tok = l.newToken(TokenOperator, "*")
	case '=':
		tok = l.newToken(TokenOperator, "=")
	case '<':
		if l.peekChar() == '>' {
			l.readChar()
			tok = l.newToken(TokenOperator, "<>")
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(TokenOperator, "<=")
		} else {
			tok = l.newToken(TokenOperator, "<")
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.newToken(TokenOperator, ">=")
		} else {
			tok = l.newToken(TokenOperator, ">")
		}
	case ',':
		tok = l.newToken(TokenOperator, ",")
	case '(':
		tok = l.newToken(TokenOperator, "(")
	case ')':
		tok = l.newToken(TokenOperator, ")")
	case '\'':
		tok.Type = TokenString
		tok.Lit = l.readString()
	case 0:
		tok.Lit = ""
		tok.Type = TokenEOF
	default:
		if isLetter(l.ch) {
			tok.Lit = l.readIdentifier()
			tok.Type = TokenIdent
			if keywords[strings.ToUpper(tok.Lit)] {
				tok.Type = TokenKeyword
				tok.Lit = strings.ToUpper(tok.Lit) // Normalizar keywords a mayúsculas
			}
			return tok
		} else if isDigit(l.ch) {
			tok.Type, tok.Lit = l.readNumber()
			return tok
		} else {
			tok = l.newToken(TokenIllegal, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) newToken(tokenType TokenType, ch string) Token {
	return Token{Type: tokenType, Lit: ch, Line: l.line, Col: l.col, Pos: l.position}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() (TokenType, string) {
	position := l.position
	isFloat := false
	for isDigit(l.ch) || l.ch == '.' {
		if l.ch == '.' {
			isFloat = true
		}
		l.readChar()
	}
	if isFloat {
		return TokenFloat, l.input[position:l.position]
	}
	return TokenInt, l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
