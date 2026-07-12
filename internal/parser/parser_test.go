package parser

import (
	"strings"
	"testing"
)

func TestParse_TableDriven(t *testing.T) {
	cases := []struct {
		name    string
		sql     string
		wantErr bool
		errMsg  string // Fragmento del error esperado para validar la posición
	}{
		{
			name:    "Consulta válida simple",
			sql:     "SELECT id, nombre FROM empleados",
			wantErr: false,
		},
		{
			name:    "Consulta válida con WHERE y operadores",
			sql:     "SELECT * FROM empleados WHERE salario > 3000 AND (activo = TRUE OR departamento = 'TI')",
			wantErr: false,
		},
		{
			name:    "Error: Falta SELECT",
			sql:     "FROM empleados",
			wantErr: true,
			errMsg:  "Línea 1, Col 1", // Verifica que detecta el error en la posición correcta
		},
		{
			name:    "Error: Sintaxis incompleta en WHERE",
			sql:     "SELECT id FROM empleados WHERE salario > ",
			wantErr: true,
			errMsg:  "se esperaba columna, valor literal o '('",
		},
		{
			name:    "Error: Basura al final de la consulta",
			sql:     "SELECT id FROM empleados LIMIT 10", // LIMIT aún no soportado por tu gramática base
			wantErr: true,
			errMsg:  "tokens inesperados al final",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := Parse(tc.sql)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("se esperaba un error, pero se obtuvo nil")
				}
				if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("el error no contiene el mensaje esperado.\nObtenido: %v\nEsperado: %s", err, tc.errMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("no se esperaba error, pero falló con: %v", err)
				}
				if ast == nil {
					t.Fatalf("se esperaba un AST válido, se obtuvo nil")
				}
			}
		})
	}
}
