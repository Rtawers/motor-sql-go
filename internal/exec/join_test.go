package exec_test

import (
	"sort"
	"testing"

	"github.com/uss-taller-go/motor-sql-go/internal/catalog"
	"github.com/uss-taller-go/motor-sql-go/internal/exec"
)

func loadDepartamentos(t *testing.T) *catalog.Table {
	t.Helper()
	tbl, err := catalog.LoadCSV("../../testdata/departamentos.csv", "departamentos")
	if err != nil {
		t.Fatalf("LoadCSV departamentos: %v", err)
	}
	return tbl
}

// nombresUbicaciones extrae pares "nombre|ubicacion" de las filas unidas,
// ordenados, para comparar resultados de forma estable.
func nombresUbicaciones(rows []exec.Row) []string {
	var out []string
	for _, r := range rows {
		nombre, _ := r.Get("empleados.nombre")
		ubic, _ := r.Get("departamentos.ubicacion")
		out = append(out, nombre.S+"|"+ubic.S)
	}
	sort.Strings(out)
	return out
}

func TestNestedLoopJoin(t *testing.T) {
	emp := loadEmpleados(t)
	dep := loadDepartamentos(t)

	join := exec.NewNestedLoopJoin(
		exec.NewScan(emp), "empleados", "departamento",
		exec.NewScan(dep), "departamentos", "depto",
	)
	rows := drain(t, join)

	// 5 empleados, todos con departamento existente (Ventas, TI, RRHH).
	// Legal no tiene empleados, así que no aparece. Esperamos 5 filas.
	if len(rows) != 5 {
		t.Fatalf("esperaba 5 filas unidas, obtuve %d", len(rows))
	}
}

func TestHashJoin(t *testing.T) {
	emp := loadEmpleados(t)
	dep := loadDepartamentos(t)

	join := exec.NewHashJoin(
		exec.NewScan(emp), "empleados", "departamento",
		exec.NewScan(dep), "departamentos", "depto",
	)
	rows := drain(t, join)
	if len(rows) != 5 {
		t.Fatalf("esperaba 5 filas unidas, obtuve %d", len(rows))
	}
}

// TestJoins_ProduceSameResult es la prueba clave que pide el enunciado:
// nested-loop y hash join deben dar EXACTAMENTE el mismo resultado.
func TestJoins_ProduceSameResult(t *testing.T) {
	emp := loadEmpleados(t)
	dep := loadDepartamentos(t)

	nl := exec.NewNestedLoopJoin(
		exec.NewScan(emp), "empleados", "departamento",
		exec.NewScan(dep), "departamentos", "depto",
	)
	hj := exec.NewHashJoin(
		exec.NewScan(emp), "empleados", "departamento",
		exec.NewScan(dep), "departamentos", "depto",
	)

	nlRows := nombresUbicaciones(drain(t, nl))
	hjRows := nombresUbicaciones(drain(t, hj))

	if len(nlRows) != len(hjRows) {
		t.Fatalf("distinto número de filas: nested=%d hash=%d", len(nlRows), len(hjRows))
	}
	for i := range nlRows {
		if nlRows[i] != hjRows[i] {
			t.Errorf("difieren en la fila %d: nested=%q hash=%q", i, nlRows[i], hjRows[i])
		}
	}
}

// TestJoin_SchemaHasPrefixedColumns verifica que el esquema combinado usa
// nombres prefijados con la tabla.
func TestJoin_SchemaHasPrefixedColumns(t *testing.T) {
	emp := loadEmpleados(t)
	dep := loadDepartamentos(t)
	join := exec.NewHashJoin(
		exec.NewScan(emp), "empleados", "departamento",
		exec.NewScan(dep), "departamentos", "depto",
	)
	schema := join.Schema()
	if schema.IndexOf("empleados.nombre") < 0 {
		t.Error("falta la columna prefijada empleados.nombre")
	}
	if schema.IndexOf("departamentos.ubicacion") < 0 {
		t.Error("falta la columna prefijada departamentos.ubicacion")
	}
}
