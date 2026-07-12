package catalog

import (
	"testing"
)

func TestLoadCSV(t *testing.T) {
	cases := []struct {
		name      string
		path      string
		tableName string
		wantErr   bool
		wantRows  int
		wantCols  int
	}{
		{
			name:      "Carga exitosa normal",
			path:      "../../testdata/empleados.csv",
			tableName: "empleados",
			wantErr:   false,
			wantRows:  5,
			wantCols:  5,
		},
		{
			name:      "Archivo con valores nulos e inferencia profunda",
			path:      "../../testdata/empleados_nulos.csv",
			tableName: "empleados_nulos",
			wantErr:   false,
			wantRows:  3,
			wantCols:  4,
		},
		{
			name:      "Archivo malformado con columnas faltantes",
			path:      "../../testdata/empleados_malos.csv",
			tableName: "empleados_malos",
			wantErr:   false, // Configurado para no fallar y rellenar con NULL
			wantRows:  2,
			wantCols:  3,
		},
		{
			name:      "Archivo inexistente lanza error",
			path:      "../../testdata/no_existe.csv",
			tableName: "nada",
			wantErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tbl, err := LoadCSV(tc.path, tc.tableName)
			if (err != nil) != tc.wantErr {
				t.Fatalf("LoadCSV() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr {
				if len(tbl.Rows) != tc.wantRows {
					t.Errorf("Esperaba %d filas, obtuve %d", tc.wantRows, len(tbl.Rows))
				}
				if len(tbl.Schema) != tc.wantCols {
					t.Errorf("Esperaba %d columnas, obtuve %d", tc.wantCols, len(tbl.Schema))
				}
			}
		})
	}
}
