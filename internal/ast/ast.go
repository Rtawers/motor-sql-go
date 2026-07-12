// Package ast define los nodos del árbol de sintaxis abstracta producido
// por internal/parser y consumido por internal/exec.
//
// PROPIEDAD: Maicol.
// Este archivo es un ESQUELETO/PUNTO DE PARTIDA, no una implementación
// completa — el diseño real de los nodos (qué campos, cómo se representa
// AND/OR anidado, etc.) es una decisión del equipo a documentar en la
// bitácora (Hito H2).

// Package ast define los nodos del árbol de sintaxis abstracta producido
// por internal/parser y consumido por internal/exec.
//
// PROPIEDAD: Maicol.
package ast

import "github.com/uss-taller-go/motor-sql-go/internal/types"

// SelectStmt representa un `SELECT ... FROM ... [WHERE ...]`.
type SelectStmt struct {
	Columns []string // "*" se representa como []string{"*"}
	From    string
	Where   Expr // nil si no hay WHERE
}

// Expr es la interfaz de cualquier nodo de expresión.
type Expr interface {
	exprNode()
}

// BinaryExpr representa operaciones binarias (AND, OR, =, <>, >, <, >=, <=).
type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryExpr) exprNode() {}

// ColumnRef representa la referencia a una columna en la consulta (ej. "salario").
type ColumnRef struct {
	Name string
}

func (c *ColumnRef) exprNode() {}

// Literal representa un valor constante (ej. 3000, 'TI', true, NULL).
type Literal struct {
	Value types.Value
}

func (l *Literal) exprNode() {}
