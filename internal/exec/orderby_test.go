package exec_test

import (
	"testing"

	"github.com/uss-taller-go/motor-sql-go/internal/exec"
)

func TestOrderBy_TableDriven(t *testing.T) {
	tbl := loadEmpleados(t)

	cases := []struct {
		name      string
		col       string
		desc      bool
		wantFirst string
		wantLast  string
	}{
		{"salario ASC", "salario", false, "Pedro", "Luis"},
		{"salario DESC", "salario", true, "Luis", "Pedro"},
		{"nombre ASC", "nombre", false, "Ana", "Sofia"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scan := exec.NewScan(tbl)
			ob := exec.NewOrderBy(scan, tc.col, tc.desc)
			rows := drain(t, ob)
			if len(rows) != 5 {
				t.Fatalf("esperaba 5 filas, obtuve %d", len(rows))
			}
			first, _ := rows[0].Get("nombre")
			last, _ := rows[len(rows)-1].Get("nombre")
			if first.S != tc.wantFirst {
				t.Errorf("primera fila: esperaba %s, obtuve %s", tc.wantFirst, first.S)
			}
			if last.S != tc.wantLast {
				t.Errorf("última fila: esperaba %s, obtuve %s", tc.wantLast, last.S)
			}
		})
	}
}

func TestOrderBy_UnknownColumn_ReturnsError(t *testing.T) {
	tbl := loadEmpleados(t)
	scan := exec.NewScan(tbl)
	ob := exec.NewOrderBy(scan, "columna_fantasma", false)
	_, _, err := ob.Next()
	if err == nil {
		t.Error("esperaba error por columna inexistente en ORDER BY")
	}
}
