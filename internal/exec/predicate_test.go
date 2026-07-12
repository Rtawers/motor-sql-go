package exec_test

import (
	"testing"

	"github.com/uss-taller-go/motor-sql-go/internal/ast"
	"github.com/uss-taller-go/motor-sql-go/internal/exec"
	"github.com/uss-taller-go/motor-sql-go/internal/parser"
	"github.com/uss-taller-go/motor-sql-go/internal/types"
)

// buildPredFromSQL parsea el WHERE de una consulta y construye su Predicate,
// usando el parser real de Maicol. Así probamos la integración completa
// parser -> traductor -> predicado.
func buildPredFromSQL(t *testing.T, sql string) exec.Predicate {
	t.Helper()
	stmt, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("parser.Parse(%q): %v", sql, err)
	}
	pred, err := exec.BuildPredicate(stmt.Where)
	if err != nil {
		t.Fatalf("BuildPredicate: %v", err)
	}
	return pred
}

func TestBuildPredicate_Integration(t *testing.T) {
	tbl := loadEmpleados(t)

	cases := []struct {
		name string
		sql  string
		want int // cuántas filas deben pasar el filtro
	}{
		{"sin where deja pasar todo", "SELECT * FROM empleados", 5},
		{"comparacion simple >", "SELECT * FROM empleados WHERE salario > 3000", 4},
		{"igualdad texto", "SELECT * FROM empleados WHERE departamento = 'TI'", 2},
		{"distinto <>", "SELECT * FROM empleados WHERE departamento <> 'TI'", 3},
		{"AND", "SELECT * FROM empleados WHERE salario > 3000 AND departamento = 'TI'", 2},
		{"OR", "SELECT * FROM empleados WHERE departamento = 'TI' OR departamento = 'RRHH'", 3},
		{"parentesis y precedencia", "SELECT * FROM empleados WHERE salario > 3000 AND (departamento = 'Ventas' OR departamento = 'RRHH')", 2},
		{"booleano", "SELECT * FROM empleados WHERE activo = TRUE", 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pred := buildPredFromSQL(t, tc.sql)
			scan := exec.NewScan(tbl)
			filter := exec.NewFilter(scan, pred)
			rows := drain(t, filter)
			if len(rows) != tc.want {
				t.Errorf("%s: esperaba %d filas, obtuve %d", tc.sql, tc.want, len(rows))
			}
		})
	}
}

// TestBuildPredicate_NullSemantics verifica la decisión de diseño: una
// comparación contra NULL descarta la fila (no la deja pasar).
func TestBuildPredicate_NullSemantics(t *testing.T) {
	// Fila con un salario NULL.
	schema := loadEmpleados(t).Schema
	row := exec.Row{
		Schema: schema,
		Values: []types.Value{types.Int(9), types.Str("Nadie"), types.Str("TI"), types.Null, types.Bool(true)},
	}

	pred, err := exec.BuildPredicate(&ast.BinaryExpr{
		Left:  &ast.ColumnRef{Name: "salario"},
		Op:    ">",
		Right: &ast.Literal{Value: types.Int(3000)},
	})
	if err != nil {
		t.Fatalf("BuildPredicate: %v", err)
	}

	pass, err := pred(row)
	if err != nil {
		t.Fatalf("pred: %v", err)
	}
	if pass {
		t.Error("una comparación contra NULL no debería dejar pasar la fila")
	}
}

// TestBuildPredicate_UnknownColumn verifica que referenciar una columna
// inexistente en el WHERE produce error (no panic).
func TestBuildPredicate_UnknownColumn(t *testing.T) {
	tbl := loadEmpleados(t)
	pred := buildPredFromSQL(t, "SELECT * FROM empleados WHERE inexistente = 1")
	scan := exec.NewScan(tbl)
	filter := exec.NewFilter(scan, pred)
	_, _, err := filter.Next()
	if err == nil {
		t.Error("esperaba error por columna inexistente en el WHERE")
	}
}
