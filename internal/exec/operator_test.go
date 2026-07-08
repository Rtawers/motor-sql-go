package exec_test

import (
	"testing"

	"github.com/uss-taller-go/motor-sql-go/internal/catalog"
	"github.com/uss-taller-go/motor-sql-go/internal/exec"
	"github.com/uss-taller-go/motor-sql-go/internal/types"
)

func loadEmpleados(t *testing.T) *catalog.Table {
	t.Helper()
	tbl, err := catalog.LoadCSV("../../testdata/empleados.csv", "empleados")
	if err != nil {
		t.Fatalf("LoadCSV: %v", err)
	}
	return tbl
}

func drain(t *testing.T, op exec.Operator) []exec.Row {
	t.Helper()
	defer op.Close()
	var rows []exec.Row
	for {
		row, ok, err := op.Next()
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if !ok {
			break
		}
		rows = append(rows, row)
	}
	return rows
}

func TestScan_ReturnsAllRows(t *testing.T) {
	tbl := loadEmpleados(t)
	scan := exec.NewScan(tbl)
	rows := drain(t, scan)
	if len(rows) != 5 {
		t.Fatalf("esperaba 5 filas, obtuve %d", len(rows))
	}
}

func TestFilter_TableDriven(t *testing.T) {
	tbl := loadEmpleados(t)

	cases := []struct {
		name string
		pred exec.Predicate
		want int
	}{
		{
			name: "departamento = TI",
			pred: func(r exec.Row) (bool, error) {
				v, _ := r.Get("departamento")
				return v.S == "TI", nil
			},
			want: 2,
		},
		{
			name: "activo = true",
			pred: func(r exec.Row) (bool, error) {
				v, _ := r.Get("activo")
				return v.Kind == types.KindBool && v.B, nil
			},
			want: 4,
		},
		{
			name: "salario > 3000 (usando types.Compare)",
			pred: func(r exec.Row) (bool, error) {
				v, _ := r.Get("salario")
				cmp, ok := types.Compare(v, types.Float(3000))
				return ok && cmp > 0, nil
			},
			want: 4,
		},
		{
			name: "ningún resultado",
			pred: func(r exec.Row) (bool, error) {
				return false, nil
			},
			want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scan := exec.NewScan(tbl)
			filter := exec.NewFilter(scan, tc.pred)
			rows := drain(t, filter)
			if len(rows) != tc.want {
				t.Errorf("%s: esperaba %d filas, obtuve %d", tc.name, tc.want, len(rows))
			}
		})
	}
}

func TestProject_SelectsColumnsInOrder(t *testing.T) {
	tbl := loadEmpleados(t)
	scan := exec.NewScan(tbl)
	proj, err := exec.NewProject(scan, []string{"nombre", "salario"})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	rows := drain(t, proj)
	if len(rows) != 5 {
		t.Fatalf("esperaba 5 filas, obtuve %d", len(rows))
	}
	first := rows[0]
	if len(first.Values) != 2 {
		t.Fatalf("esperaba 2 columnas proyectadas, obtuve %d", len(first.Values))
	}
	if first.Schema[0].Name != "nombre" || first.Schema[1].Name != "salario" {
		t.Errorf("orden de columnas incorrecto: %v", first.Schema)
	}
}

func TestProject_UnknownColumn_ReturnsError(t *testing.T) {
	tbl := loadEmpleados(t)
	scan := exec.NewScan(tbl)
	_, err := exec.NewProject(scan, []string{"columna_inexistente"})
	if err == nil {
		t.Fatal("esperaba error por columna inexistente, obtuve nil")
	}
}

func TestOperatorChain_ScanFilterProject(t *testing.T) {
	// Demuestra la composición: agregar Project sobre Filter sobre Scan
	// sin que ninguno conozca la existencia del otro.
	tbl := loadEmpleados(t)
	scan := exec.NewScan(tbl)
	filter := exec.NewFilter(scan, func(r exec.Row) (bool, error) {
		v, _ := r.Get("departamento")
		return v.S == "Ventas", nil
	})
	proj, err := exec.NewProject(filter, []string{"nombre"})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	rows := drain(t, proj)
	if len(rows) != 2 {
		t.Fatalf("esperaba 2 vendedores, obtuve %d", len(rows))
	}
}
