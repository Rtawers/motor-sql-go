// Package parser construye un internal/ast.SelectStmt a partir de tokens de
// internal/lexer, usando descenso recursivo (o Pratt para expresiones).
//
// PROPIEDAD: Maicol.
package parser

import (
	"github.com/uss-taller-go/motor-sql-go/internal/ast"
)

// Parse analiza el texto SQL de entrada y devuelve el AST de un SELECT, o un
// error de sintaxis con mensaje y posición (requisito de rúbrica, 20%).
//
// TODO(maicol, H2): implementar de verdad. Firma sujeta a ajuste del equipo.
func Parse(sql string) (*ast.SelectStmt, error) {
	panic("parser.Parse: no implementado todavía — ver docs/gramatica.md")
}
