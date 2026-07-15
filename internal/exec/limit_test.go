package exec_test

import (
	"testing"

	"github.com/uss-taller-go/motor-sql-go/internal/exec"
)

func TestLimit_CutsRowsCorrectly(t *testing.T) {
	tbl := loadEmpleados(t)
	scan := exec.NewScan(tbl)
	limit := exec.NewLimit(scan, 2)
	
	rows := drain(t, limit)
	if len(rows) != 2 {
		t.Errorf("esperaba 2 filas, obtuve %d", len(rows))
	}
	
	first, _ := rows[0].Get("id")
	if first.I != 1 {
		t.Errorf("esperaba id 1 en la primera fila, obtuve %d", first.I)
	}
}