package exec_test

import (
	"testing"

	"github.com/uss-taller-go/motor-sql-go/internal/exec"
)

func TestGroupBy_TableDriven(t *testing.T) {
	tbl := loadEmpleados(t)

	t.Run("Agrupacion global: COUNT(*)", func(t *testing.T) {
		scan := exec.NewScan(tbl)
		aggs := []exec.AggregateDef{{Func: "COUNT", Column: "*"}}
		gb, _ := exec.NewGroupBy(scan, "", aggs)

		rows := drain(t, gb)
		if len(rows) != 1 {
			t.Fatalf("esperaba 1 fila global, obtuve %d", len(rows))
		}

		val, _ := rows[0].Get("COUNT(*)")
		if val.I != 5 {
			t.Errorf("esperaba COUNT(*) = 5, obtuve %d", val.I)
		}
	})

	t.Run("Agrupacion por departamento", func(t *testing.T) {
		scan := exec.NewScan(tbl)
		aggs := []exec.AggregateDef{{Func: "COUNT", Column: "*"}}
		gb, _ := exec.NewGroupBy(scan, "departamento", aggs)

		rows := drain(t, gb)
		// testdata/empleados.csv tiene 3 departamentos: Ventas, TI, RRHH
		if len(rows) != 3 {
			t.Fatalf("esperaba 3 grupos, obtuve %d", len(rows))
		}
	})
}
