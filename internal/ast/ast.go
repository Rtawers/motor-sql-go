// Package ast define los nodos del árbol de sintaxis abstracta producido
// por internal/parser y consumido por internal/exec.
//
// PROPIEDAD: Maicol.
// Este archivo es un ESQUELETO/PUNTO DE PARTIDA, no una implementación
// completa — el diseño real de los nodos (qué campos, cómo se representa
// AND/OR anidado, etc.) es una decisión del equipo a documentar en la
// bitácora (Hito H2).
package ast

// SelectStmt representa un `SELECT ... FROM ... [WHERE ...] [ORDER BY ...]
// [LIMIT ...]`. Completar/ajustar campos según se implementen los hitos.
type SelectStmt struct {
	Columns []string // "*" se representa como []string{"*"}
	From    string
	Where   Expr // nil si no hay WHERE
	// TODO(H4): OrderBy, Limit, GroupBy, Aggregates
	// TODO(H5): Join
}

// Expr es la interfaz de cualquier nodo de expresión (WHERE, ON de join).
type Expr interface {
	exprNode()
}

// TODO(maicol, H2): definir los nodos concretos, por ejemplo:
//   - BinaryExpr{Left, Op, Right} para comparaciones y AND/OR
//   - ColumnRef{Name}
//   - Literal{Value}
// y hacer que cada uno implemente exprNode().
