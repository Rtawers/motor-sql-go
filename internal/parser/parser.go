// Package parser construye un internal/ast.SelectStmt a partir de tokens de
// internal/lexer, usando descenso recursivo (o Pratt para expresiones).
//
// PROPIEDAD: Maicol.

// Package parser construye un internal/ast.SelectStmt a partir de tokens de
// internal/lexer, usando descenso recursivo.
//
// PROPIEDAD: Maicol.
package parser

import (
	"fmt"
	"strconv"

	"github.com/uss-taller-go/motor-sql-go/internal/ast"
	"github.com/uss-taller-go/motor-sql-go/internal/lexer"
	"github.com/uss-taller-go/motor-sql-go/internal/types"
)

type Parser struct {
	l    *lexer.Lexer
	cur  lexer.Token
	peek lexer.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// Leer los dos primeros tokens para inicializar cur y peek
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.l.NextToken()
}

// Parse analiza el texto SQL y devuelve el AST.
func Parse(sql string) (*ast.SelectStmt, error) {
	l := lexer.New(sql)
	p := New(l)
	return p.parseSelectStmt()
}

func (p *Parser) errorMsg(msg string) error {
	return fmt.Errorf("error de sintaxis [Línea %d, Col %d]: %s (obtenido: '%s')", p.cur.Line, p.cur.Col, msg, p.cur.Lit)
}

func (p *Parser) parseSelectStmt() (*ast.SelectStmt, error) {
	stmt := &ast.SelectStmt{}

	// 1. SELECT
	if p.cur.Type != lexer.TokenKeyword || p.cur.Lit != "SELECT" {
		return nil, p.errorMsg("se esperaba SELECT")
	}
	p.nextToken()

	// 2. Columnas
	for {
		if p.cur.Type == lexer.TokenOperator && p.cur.Lit == "*" {
			stmt.Columns = append(stmt.Columns, "*")
			p.nextToken()
		} else if p.cur.Type == lexer.TokenIdent {
			stmt.Columns = append(stmt.Columns, p.cur.Lit)
			p.nextToken()
		} else {
			return nil, p.errorMsg("se esperaba columna o '*'")
		}

		if p.cur.Type == lexer.TokenOperator && p.cur.Lit == "," {
			p.nextToken() // saltar la coma
		} else {
			break
		}
	}

	// 3. FROM
	if p.cur.Type != lexer.TokenKeyword || p.cur.Lit != "FROM" {
		return nil, p.errorMsg("se esperaba FROM")
	}
	p.nextToken()

	// 4. Tabla
	if p.cur.Type != lexer.TokenIdent {
		return nil, p.errorMsg("se esperaba nombre de tabla")
	}
	stmt.From = p.cur.Lit
	p.nextToken()

	// 5. WHERE (opcional)
	if p.cur.Type == lexer.TokenKeyword && p.cur.Lit == "WHERE" {
		p.nextToken()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Where = expr
	}

	// Asegurar que no haya basura al final
	if p.cur.Type != lexer.TokenEOF {
		return nil, p.errorMsg("tokens inesperados al final de la consulta")
	}

	return stmt, nil
}

// ---- Evaluación de Expresiones (Descenso Recursivo) ----

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseOrExpr()
}

func (p *Parser) parseOrExpr() (ast.Expr, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for p.cur.Type == lexer.TokenKeyword && p.cur.Lit == "OR" {
		op := p.cur.Lit
		p.nextToken()
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAndExpr() (ast.Expr, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.cur.Type == lexer.TokenKeyword && p.cur.Lit == "AND" {
		op := p.cur.Lit
		p.nextToken()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseComparison() (ast.Expr, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	if p.cur.Type == lexer.TokenOperator && isComparisonOp(p.cur.Lit) {
		op := p.cur.Lit
		p.nextToken()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpr{Left: left, Op: op, Right: right}, nil
	}

	return left, nil
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	switch p.cur.Type {
	case lexer.TokenIdent:
		col := &ast.ColumnRef{Name: p.cur.Lit}
		p.nextToken()
		return col, nil
	case lexer.TokenInt:
		val, _ := strconv.ParseInt(p.cur.Lit, 10, 64)
		lit := &ast.Literal{Value: types.Int(val)}
		p.nextToken()
		return lit, nil
	case lexer.TokenFloat:
		val, _ := strconv.ParseFloat(p.cur.Lit, 64)
		lit := &ast.Literal{Value: types.Float(val)}
		p.nextToken()
		return lit, nil
	case lexer.TokenString:
		lit := &ast.Literal{Value: types.Str(p.cur.Lit)}
		p.nextToken()
		return lit, nil
	case lexer.TokenKeyword:
		if p.cur.Lit == "TRUE" {
			p.nextToken()
			return &ast.Literal{Value: types.Bool(true)}, nil
		} else if p.cur.Lit == "FALSE" {
			p.nextToken()
			return &ast.Literal{Value: types.Bool(false)}, nil
		} else if p.cur.Lit == "NULL" {
			p.nextToken()
			return &ast.Literal{Value: types.Null}, nil
		}
	case lexer.TokenOperator:
		if p.cur.Lit == "(" {
			p.nextToken()
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			if p.cur.Type != lexer.TokenOperator || p.cur.Lit != ")" {
				return nil, p.errorMsg("se esperaba ')'")
			}
			p.nextToken()
			return expr, nil
		}
	}

	return nil, p.errorMsg("se esperaba columna, valor literal o '('")
}

func isComparisonOp(op string) bool {
	return op == "=" || op == "<>" || op == "<" || op == ">" || op == "<=" || op == ">="
}
