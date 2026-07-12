// Este archivo es el PUENTE entre el parser (Maicol) y el motor de
// ejecución. Traduce una expresión del AST (ast.Expr) en un Predicate que
// el operador Filter sabe evaluar fila por fila.
//
// PROPIEDAD: Daniel + Yokt.
//
// Idea central: una expresión SQL como
//
//	salario > 3000 AND departamento = 'TI'
//
// llega como un árbol de ast.BinaryExpr. Lo recorremos recursivamente y
// evaluamos cada nodo contra la fila actual, produciendo un types.Value
// booleano. BuildPredicate envuelve esa evaluación en la firma Predicate.
package exec

import (
	"fmt"

	"github.com/uss-taller-go/motor-sql-go/internal/ast"
	"github.com/uss-taller-go/motor-sql-go/internal/types"
)

// BuildPredicate traduce una expresión del WHERE (ast.Expr) en un Predicate
// que Filter puede usar. Si expr es nil (no había WHERE), devuelve un
// predicado que deja pasar todas las filas.
func BuildPredicate(expr ast.Expr) (Predicate, error) {
	if expr == nil {
		return func(Row) (bool, error) { return true, nil }, nil
	}
	return func(row Row) (bool, error) {
		v, err := evalExpr(expr, row)
		if err != nil {
			return false, err
		}
		// El WHERE solo deja pasar filas cuyo resultado sea booleano true.
		// NULL o cualquier no-booleano se trata como "no pasa" (coherente
		// con el estándar SQL, donde un WHERE que evalúa a NULL descarta la
		// fila).
		return v.Kind == types.KindBool && v.B, nil
	}, nil
}

// evalExpr evalúa una expresión del AST contra una fila y devuelve su valor.
// Es recursiva: un BinaryExpr evalúa sus dos lados y los combina.
func evalExpr(expr ast.Expr, row Row) (types.Value, error) {
	switch e := expr.(type) {

	case *ast.Literal:
		// Un literal (3000, 'TI', true, NULL) se evalúa a sí mismo.
		return e.Value, nil

	case *ast.ColumnRef:
		// Una columna se resuelve buscando su valor en la fila.
		v, ok := row.Get(e.Name)
		if !ok {
			return types.Null, fmt.Errorf("exec: columna %q no existe", e.Name)
		}
		return v, nil

	case *ast.BinaryExpr:
		return evalBinary(e, row)

	default:
		return types.Null, fmt.Errorf("exec: tipo de expresión no soportado: %T", expr)
	}
}

// evalBinary evalúa un nodo binario. Separa dos familias de operadores:
//   - lógicos (AND, OR): combinan dos booleanos.
//   - de comparación (=, <>, <, >, <=, >=): comparan dos valores.
func evalBinary(e *ast.BinaryExpr, row Row) (types.Value, error) {
	switch e.Op {
	case "AND", "OR":
		return evalLogical(e, row)
	case "=", "<>", "<", ">", "<=", ">=":
		return evalComparison(e, row)
	default:
		return types.Null, fmt.Errorf("exec: operador no soportado: %q", e.Op)
	}
}

func evalLogical(e *ast.BinaryExpr, row Row) (types.Value, error) {
	left, err := evalExpr(e.Left, row)
	if err != nil {
		return types.Null, err
	}
	right, err := evalExpr(e.Right, row)
	if err != nil {
		return types.Null, err
	}
	if left.Kind != types.KindBool || right.Kind != types.KindBool {
		return types.Null, fmt.Errorf("exec: %s requiere operandos booleanos", e.Op)
	}
	if e.Op == "AND" {
		return types.Bool(left.B && right.B), nil
	}
	return types.Bool(left.B || right.B), nil
}

func evalComparison(e *ast.BinaryExpr, row Row) (types.Value, error) {
	left, err := evalExpr(e.Left, row)
	if err != nil {
		return types.Null, err
	}
	right, err := evalExpr(e.Right, row)
	if err != nil {
		return types.Null, err
	}

	// DECISIÓN DE DISEÑO (documentar en bitácora): cualquier comparación con
	// NULL da NULL (no true, no false), siguiendo la semántica SQL. Como el
	// WHERE solo deja pasar true, una fila con NULL en la comparación se
	// descarta. types.Compare ya devuelve ok=false cuando hay NULL.
	cmp, ok := types.Compare(left, right)
	if !ok {
		return types.Null, nil
	}

	switch e.Op {
	case "=":
		return types.Bool(cmp == 0), nil
	case "<>":
		return types.Bool(cmp != 0), nil
	case "<":
		return types.Bool(cmp < 0), nil
	case ">":
		return types.Bool(cmp > 0), nil
	case "<=":
		return types.Bool(cmp <= 0), nil
	case ">=":
		return types.Bool(cmp >= 0), nil
	default:
		return types.Null, fmt.Errorf("exec: operador de comparación no soportado: %q", e.Op)
	}
}
