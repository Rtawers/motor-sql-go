// Package catalog implementa H1: carga de archivos CSV a tablas en memoria
// con un catálogo consultable de tablas y esquemas.
//
// PROPIEDAD: Pamela.
// TODO(pamela): este archivo es un punto de partida mínimo, no la solución
// completa. Falta al menos:
//   - Decidir inferencia de tipos vs. declaración explícita (documentar en
//     bitácora H1) — ahora mismo InferType es una heurística simple.
//   - Manejo de CSV con comillas/escapes raros, columnas faltantes por fila.
//   - Tests de tabla (table-driven) cubriendo tipos mixtos, NULL, errores.

// Package catalog implementa H1: carga de archivos CSV a tablas en memoria
// con un catálogo consultable de tablas y esquemas.
//
// PROPIEDAD: Pamela.
package catalog

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/uss-taller-go/motor-sql-go/internal/types"
)

// Column describe una columna de una tabla.
type Column struct {
	Name string
	Kind types.Kind
}

// Schema es la lista ordenada de columnas de una tabla.
type Schema []Column

// IndexOf devuelve la posición de una columna por nombre, o -1 si no existe.
func (s Schema) IndexOf(name string) int {
	for i, c := range s {
		if c.Name == name {
			return i
		}
	}
	return -1
}

// Table es una tabla en memoria: esquema + filas de valores.
type Table struct {
	Name   string
	Schema Schema
	Rows   [][]types.Value
}

// Catalog registra las tablas disponibles, consultable por nombre.
type Catalog struct {
	tables map[string]*Table
}

func NewCatalog() *Catalog {
	return &Catalog{tables: make(map[string]*Table)}
}

func (c *Catalog) Register(t *Table) {
	c.tables[t.Name] = t
}

// Table busca una tabla por nombre. ok=false si no existe.
func (c *Catalog) Table(name string) (*Table, bool) {
	t, ok := c.tables[name]
	return t, ok
}

func (c *Catalog) TableNames() []string {
	names := make([]string, 0, len(c.tables))
	for name := range c.tables {
		names = append(names, name)
	}
	return names
}

// LoadCSV carga un archivo CSV como una tabla nombrada `tableName`.
// La primera fila del CSV se toma como cabecera (nombres de columna).
// El tipo de cada columna se infiere escaneando el primer dato válido.
func LoadCSV(path, tableName string) (*Table, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("catalog: no se pudo abrir %q: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1    // Permite filas con distinto número de columnas sin lanzar error
	r.TrimLeadingSpace = true // Limpia espacios en blanco innecesarios

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("catalog: error leyendo CSV %q: %w", path, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("catalog: %q está vacío", path)
	}

	header := records[0]
	dataRows := records[1:]

	schema := make(Schema, len(header))
	for i, name := range header {
		kind := types.KindString

		// Inferencia mejorada: escanea hacia abajo hasta encontrar un valor no vacío
		for _, row := range dataRows {
			if i < len(row) && row[i] != "" {
				kind = InferType(row[i])
				break
			}
		}
		schema[i] = Column{Name: name, Kind: kind}
	}

	rows := make([][]types.Value, 0, len(dataRows))
	for _, rec := range dataRows {
		row := make([]types.Value, len(schema))
		for i := range schema {
			raw := ""
			// Manejo seguro si la fila actual tiene menos columnas que la cabecera
			if i < len(rec) {
				raw = rec[i]
			}
			row[i] = ParseValue(raw, schema[i].Kind)
		}
		rows = append(rows, row)
	}

	return &Table{Name: tableName, Schema: schema, Rows: rows}, nil
}

// InferType es una heurística MÍNIMA de inferencia de tipo a partir de un string crudo.
func InferType(raw string) types.Kind {
	if raw == "" {
		return types.KindString
	}
	if _, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return types.KindInt
	}
	if _, err := strconv.ParseFloat(raw, 64); err == nil {
		return types.KindFloat
	}
	if raw == "true" || raw == "false" {
		return types.KindBool
	}
	return types.KindString
}

// ParseValue convierte un string crudo a un types.Value según el Kind dado.
// Un string vacío se interpreta como NULL.
func ParseValue(raw string, kind types.Kind) types.Value {
	if raw == "" {
		return types.Null
	}
	switch kind {
	case types.KindInt:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return types.Str(raw) // fallback: no debería pasar si InferType fue consistente
		}
		return types.Int(v)
	case types.KindFloat:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return types.Str(raw)
		}
		return types.Float(v)
	case types.KindBool:
		return types.Bool(raw == "true")
	default:
		return types.Str(raw)
	}
}
