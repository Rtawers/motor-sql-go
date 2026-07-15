// Package ast define los nodos del árbol de sintaxis abstracta producido
// por internal/parser y consumido por internal/exec.
//
// PROPIEDAD: Maicol.
package ast

import "github.com/uss-taller-go/motor-sql-go/internal/types"

// SelectStmt representa un `SELECT ... FROM ... [JOIN ...] [WHERE ...] [GROUP BY ...] [ORDER BY ...] [LIMIT ...]`.
type SelectStmt struct {
	SelectItems []SelectItem
	From        string
	Join        *JoinClause // H5: Lennart
	Where       Expr        // nil si no hay WHERE
	GroupBy     []string    // H4: Victoria
	OrderBy     []OrderItem // H4: Victoria
	Limit       *int        // H4: Victoria (puntero para saber si es nulo)
}

// ---- H4: Elementos del SELECT ----

// SelectItem es la interfaz para columnas simples o funciones de agregación.
type SelectItem interface {
	selectItemNode()
}

// ColumnItem representa una columna normal o un "*".
type ColumnItem struct {
	Name string
}

func (c *ColumnItem) selectItemNode() {}

// AggregateItem representa COUNT(col), SUM(col), etc.
type AggregateItem struct {
	Func   string // COUNT, SUM, AVG, MIN, MAX
	Column string // Nombre de la columna o "*"
}

func (a *AggregateItem) selectItemNode() {}

// ---- H4: ORDER BY ----

// OrderItem representa una columna y su dirección de ordenamiento.
type OrderItem struct {
	Column string
	Desc   bool // false = ASC, true = DESC
}

// ---- H5: JOIN ----

// JoinClause representa un INNER JOIN ... ON ...
type JoinClause struct {
	Table string
	On    Expr
}

// ---- Expresiones (Ya implementado) ----

type Expr interface {
	exprNode()
}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryExpr) exprNode() {}

type ColumnRef struct {
	Name string
}

func (c *ColumnRef) exprNode() {}

type Literal struct {
	Value types.Value
}

func (l *Literal) exprNode() {}
