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

func isAggregateFunc(lit string) bool {
	return lit == "COUNT" || lit == "SUM" || lit == "AVG" || lit == "MIN" || lit == "MAX"
}

func (p *Parser) parseSelectStmt() (*ast.SelectStmt, error) {
	stmt := &ast.SelectStmt{}

	// 1. SELECT
	if p.cur.Type != lexer.TokenKeyword || p.cur.Lit != "SELECT" {
		return nil, p.errorMsg("se esperaba SELECT")
	}
	p.nextToken()

	// 2. Columnas y Agregados (H4)
	for {
		if p.cur.Type == lexer.TokenOperator && p.cur.Lit == "*" {
			stmt.SelectItems = append(stmt.SelectItems, &ast.ColumnItem{Name: "*"})
			p.nextToken()
		} else if p.cur.Type == lexer.TokenKeyword && isAggregateFunc(p.cur.Lit) {
			funcName := p.cur.Lit
			p.nextToken()
			if p.cur.Lit != "(" {
				return nil, p.errorMsg("se esperaba '(' después de la función de agregación")
			}
			p.nextToken()
			if p.cur.Type != lexer.TokenIdent && p.cur.Lit != "*" {
				return nil, p.errorMsg("se esperaba columna o '*' dentro de la función")
			}
			colName := p.cur.Lit
			p.nextToken()
			if p.cur.Lit != ")" {
				return nil, p.errorMsg("se esperaba ')' cerrando la función")
			}
			p.nextToken()
			stmt.SelectItems = append(stmt.SelectItems, &ast.AggregateItem{Func: funcName, Column: colName})
		} else if p.cur.Type == lexer.TokenIdent {
			stmt.SelectItems = append(stmt.SelectItems, &ast.ColumnItem{Name: p.cur.Lit})
			p.nextToken()
		} else {
			return nil, p.errorMsg("se esperaba columna, '*' o función de agregación")
		}

		if p.cur.Type == lexer.TokenOperator && p.cur.Lit == "," {
			p.nextToken()
		} else {
			break
		}
	}

	// 3. FROM
	if p.cur.Type != lexer.TokenKeyword || p.cur.Lit != "FROM" {
		return nil, p.errorMsg("se esperaba FROM")
	}
	p.nextToken()

	if p.cur.Type != lexer.TokenIdent {
		return nil, p.errorMsg("se esperaba nombre de tabla")
	}
	stmt.From = p.cur.Lit
	p.nextToken()

	// 4. INNER JOIN (H5)
	if p.cur.Type == lexer.TokenKeyword && p.cur.Lit == "INNER" {
		p.nextToken()
		if p.cur.Lit != "JOIN" {
			return nil, p.errorMsg("se esperaba JOIN")
		}
		p.nextToken()
		if p.cur.Type != lexer.TokenIdent {
			return nil, p.errorMsg("se esperaba nombre de tabla para el JOIN")
		}
		joinTable := p.cur.Lit
		p.nextToken()
		if p.cur.Lit != "ON" {
			return nil, p.errorMsg("se esperaba ON")
		}
		p.nextToken()
		joinExpr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Join = &ast.JoinClause{Table: joinTable, On: joinExpr}
	}

	// 5. WHERE
	if p.cur.Type == lexer.TokenKeyword && p.cur.Lit == "WHERE" {
		p.nextToken()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Where = expr
	}

	// 6. GROUP BY (H4)
	if p.cur.Type == lexer.TokenKeyword && p.cur.Lit == "GROUP" {
		p.nextToken()
		if p.cur.Lit != "BY" {
			return nil, p.errorMsg("se esperaba BY")
		}
		p.nextToken()
		for {
			if p.cur.Type != lexer.TokenIdent {
				return nil, p.errorMsg("se esperaba nombre de columna en GROUP BY")
			}
			stmt.GroupBy = append(stmt.GroupBy, p.cur.Lit)
			p.nextToken()
			if p.cur.Lit == "," {
				p.nextToken()
			} else {
				break
			}
		}
	}

	// 7. ORDER BY (H4)
	if p.cur.Type == lexer.TokenKeyword && p.cur.Lit == "ORDER" {
		p.nextToken()
		if p.cur.Lit != "BY" {
			return nil, p.errorMsg("se esperaba BY")
		}
		p.nextToken()
		for {
			if p.cur.Type != lexer.TokenIdent {
				return nil, p.errorMsg("se esperaba nombre de columna en ORDER BY")
			}
			colName := p.cur.Lit
			p.nextToken()

			desc := false
			if p.cur.Type == lexer.TokenKeyword && (p.cur.Lit == "ASC" || p.cur.Lit == "DESC") {
				desc = (p.cur.Lit == "DESC")
				p.nextToken()
			}
			stmt.OrderBy = append(stmt.OrderBy, ast.OrderItem{Column: colName, Desc: desc})

			if p.cur.Lit == "," {
				p.nextToken()
			} else {
				break
			}
		}
	}

	// 8. LIMIT (H4)
	if p.cur.Type == lexer.TokenKeyword && p.cur.Lit == "LIMIT" {
		p.nextToken()
		if p.cur.Type != lexer.TokenInt {
			return nil, p.errorMsg("se esperaba un número para el LIMIT")
		}
		limitVal, _ := strconv.Atoi(p.cur.Lit)
		stmt.Limit = &limitVal
		p.nextToken()
	}

	if p.cur.Type != lexer.TokenEOF {
		return nil, p.errorMsg("tokens inesperados al final de la consulta")
	}

	return stmt, nil
}

// ---- Evaluación de Expresiones (Sin cambios) ----

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
