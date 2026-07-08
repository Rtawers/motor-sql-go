// Package exec implementa el árbol de ejecución por operadores (modelo
// Volcano/iterador). Es el paquete de mayor peso en la rúbrica (25%): la
// interfaz debe permitir agregar operadores nuevos SIN modificar los
// existentes (abierto/cerrado).
//
// PROPIEDAD: Daniel (coordinador) para el núcleo (Operator, Scan, Filter,
// Project). Victoria añade OrderBy/Limit/GroupBy+agregados sobre esta base.
// Lennart añade los Join sobre esta misma interfaz.
package exec

import (
	"fmt"

	"github.com/uss-taller-go/motor-sql-go/internal/catalog"
	"github.com/uss-taller-go/motor-sql-go/internal/types"
)

// Row es una fila de resultado: valores posicionales + referencia a su
// esquema para poder resolver nombres de columna.
type Row struct {
	Schema catalog.Schema
	Values []types.Value
}

// Get devuelve el valor de una columna por nombre. ok=false si no existe.
func (r Row) Get(col string) (types.Value, bool) {
	idx := r.Schema.IndexOf(col)
	if idx < 0 {
		return types.Null, false
	}
	return r.Values[idx], true
}

// Operator es la interfaz común de TODOS los nodos del árbol de ejecución.
// Modelo pull/iterador: cada llamada a Next() produce una fila o señala fin.
//
// DECISIÓN DE DISEÑO CLAVE (documentar en bitácora, criterio de 25%):
// Next() devuelve (Row, bool, error) en vez de usar un sentinel tipo io.EOF.
// Evalúen en equipo si prefieren la convención io.EOF (más idiomática en Go
// estándar) y justifiquen la elección — cualquiera de las dos es válida
// mientras sea consistente en TODO el árbol.
type Operator interface {
	// Next devuelve la siguiente fila. ok=false indica fin de datos (no
	// error). Un error != nil siempre implica ok=false.
	Next() (row Row, ok bool, err error)

	// Schema describe las columnas que produce este operador.
	Schema() catalog.Schema

	// Close libera recursos (cierra archivos, limpia buffers). Debe ser
	// idempotente y llamable aunque Next() no se haya agotado.
	Close() error
}

// ---------------------------------------------------------------------
// Scan: operador hoja que recorre una Table del catálogo fila por fila.
// ---------------------------------------------------------------------

type ScanOp struct {
	table *catalog.Table
	pos   int
}

func NewScan(t *catalog.Table) *ScanOp {
	return &ScanOp{table: t}
}

func (s *ScanOp) Schema() catalog.Schema { return s.table.Schema }

func (s *ScanOp) Next() (Row, bool, error) {
	if s.pos >= len(s.table.Rows) {
		return Row{}, false, nil
	}
	row := Row{Schema: s.table.Schema, Values: s.table.Rows[s.pos]}
	s.pos++
	return row, true, nil
}

func (s *ScanOp) Close() error { return nil }

// ---------------------------------------------------------------------
// Filter: operador unario que descarta filas según un predicado.
// ---------------------------------------------------------------------

// Predicate evalúa una fila y devuelve si debe pasar el filtro.
// TODO(equipo, H2/H3): reemplazar por evaluación real de expresiones del AST
// (comparaciones, AND/OR). Esta func-type es el punto de integración: el
// parser produce un AST de expresión, y alguien debe escribir un
// "evaluador" que compile/interprete ese AST a un Predicate (o lo evalúe
// directamente fila por fila).
type Predicate func(Row) (bool, error)

type FilterOp struct {
	child Operator
	pred  Predicate
}

func NewFilter(child Operator, pred Predicate) *FilterOp {
	return &FilterOp{child: child, pred: pred}
}

func (f *FilterOp) Schema() catalog.Schema { return f.child.Schema() }

func (f *FilterOp) Next() (Row, bool, error) {
	for {
		row, ok, err := f.child.Next()
		if err != nil || !ok {
			return Row{}, false, err
		}
		pass, err := f.pred(row)
		if err != nil {
			return Row{}, false, err
		}
		if pass {
			return row, true, nil
		}
		// no pasó el filtro: seguimos consumiendo del hijo (evaluación
		// perezosa fila a fila, no se materializa nada).
	}
}

func (f *FilterOp) Close() error { return f.child.Close() }

// ---------------------------------------------------------------------
// Project: operador unario que selecciona/reordena columnas.
// ---------------------------------------------------------------------

type ProjectOp struct {
	child   Operator
	cols    []string
	schema  catalog.Schema
	indices []int
}

// NewProject construye un Project sobre las columnas indicadas por nombre.
// Devuelve error si alguna columna no existe en el esquema del hijo (no
// panic: requisito de la rúbrica de manejo de errores).
func NewProject(child Operator, cols []string) (*ProjectOp, error) {
	childSchema := child.Schema()
	schema := make(catalog.Schema, len(cols))
	indices := make([]int, len(cols))
	for i, name := range cols {
		idx := childSchema.IndexOf(name)
		if idx < 0 {
			return nil, fmt.Errorf("exec: columna %q no existe", name)
		}
		schema[i] = childSchema[idx]
		indices[i] = idx
	}
	return &ProjectOp{child: child, cols: cols, schema: schema, indices: indices}, nil
}

func (p *ProjectOp) Schema() catalog.Schema { return p.schema }

func (p *ProjectOp) Next() (Row, bool, error) {
	row, ok, err := p.child.Next()
	if err != nil || !ok {
		return Row{}, false, err
	}
	values := make([]types.Value, len(p.indices))
	for i, idx := range p.indices {
		values[i] = row.Values[idx]
	}
	return Row{Schema: p.schema, Values: values}, true, nil
}

func (p *ProjectOp) Close() error { return p.child.Close() }
