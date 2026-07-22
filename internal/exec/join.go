// Operadores de JOIN (H5). Implementa INNER JOIN de dos formas para poder
// compararlas, como pide el enunciado:
//
//   - NestedLoopJoin: por cada fila de la izquierda, recorre TODA la derecha
//     buscando coincidencias. Simple, no necesita memoria extra, pero cuesta
//     O(N*M): lento si ambas tablas son grandes.
//
//   - HashJoin: primero construye una tabla hash de la relación derecha
//     (indexada por la columna de join), luego recorre la izquierda una vez
//     y busca en el hash en O(1). Cuesta O(N+M): mucho más rápido, a cambio
//     de guardar la derecha en memoria.
//
// PROPIEDAD: Lennart.
//
// Ambos producen filas COMBINADAS: el esquema resultante concatena las
// columnas de la izquierda y la derecha, prefijadas con el nombre de su
// tabla (ej. "empleados.id", "departamentos.depto") para evitar colisiones
// cuando ambas tablas tienen columnas con el mismo nombre.
package exec

import (
	"github.com/uss-taller-go/motor-sql-go/internal/catalog"
	"github.com/uss-taller-go/motor-sql-go/internal/types"
)

// joinedSchema construye el esquema combinado, prefijando cada columna con
// el nombre de su tabla. Compartido por ambos joins.
func joinedSchema(left catalog.Schema, leftTable string, right catalog.Schema, rightTable string) catalog.Schema {
	out := make(catalog.Schema, 0, len(left)+len(right))
	for _, c := range left {
		out = append(out, catalog.Column{Name: leftTable + "." + c.Name, Kind: c.Kind})
	}
	for _, c := range right {
		out = append(out, catalog.Column{Name: rightTable + "." + c.Name, Kind: c.Kind})
	}
	return out
}

// joinRows concatena los valores de dos filas en una sola.
func joinRows(schema catalog.Schema, left, right Row) Row {
	values := make([]types.Value, 0, len(left.Values)+len(right.Values))
	values = append(values, left.Values...)
	values = append(values, right.Values...)
	return Row{Schema: schema, Values: values}
}

// ---------------------------------------------------------------------
// NestedLoopJoin
// ---------------------------------------------------------------------

type NestedLoopJoinOp struct {
	left, right           Operator
	leftTable, rightTable string
	leftCol, rightCol     string // columnas de la condición ON (left.leftCol = right.rightCol)

	schema      catalog.Schema
	rightBuf    []Row // la derecha se materializa una vez para poder re-recorrerla
	curLeft     Row
	haveLeft    bool
	rightPos    int
	initialized bool
	leftIdx     int
	rightIdx    int
}

// NewNestedLoopJoin crea un INNER JOIN por nested-loop.
// La condición es: leftTable.leftCol = rightTable.rightCol.
func NewNestedLoopJoin(left Operator, leftTable, leftCol string, right Operator, rightTable, rightCol string) *NestedLoopJoinOp {
	return &NestedLoopJoinOp{
		left: left, right: right,
		leftTable: leftTable, rightTable: rightTable,
		leftCol: leftCol, rightCol: rightCol,
		schema: joinedSchema(left.Schema(), leftTable, right.Schema(), rightTable),
	}
}

func (j *NestedLoopJoinOp) Schema() catalog.Schema { return j.schema }

func (j *NestedLoopJoinOp) init() error {
	// Materializamos la derecha (hay que recorrerla varias veces).
	for {
		row, ok, err := j.right.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		j.rightBuf = append(j.rightBuf, row)
	}
	j.leftIdx = j.left.Schema().IndexOf(j.leftCol)
	if len(j.rightBuf) > 0 {
		j.rightIdx = j.rightBuf[0].Schema.IndexOf(j.rightCol)
	}
	j.initialized = true
	return nil
}

func (j *NestedLoopJoinOp) Next() (Row, bool, error) {
	if !j.initialized {
		if err := j.init(); err != nil {
			return Row{}, false, err
		}
	}
	for {
		// Si no tenemos fila izquierda activa, pedimos la siguiente.
		if !j.haveLeft {
			row, ok, err := j.left.Next()
			if err != nil || !ok {
				return Row{}, false, err
			}
			j.curLeft = row
			j.haveLeft = true
			j.rightPos = 0
		}
		// Recorremos la derecha buscando coincidencia con la izquierda actual.
		for j.rightPos < len(j.rightBuf) {
			r := j.rightBuf[j.rightPos]
			j.rightPos++
			lv := j.curLeft.Values[j.leftIdx]
			rv := r.Values[j.rightIdx]
			if cmp, ok := types.Compare(lv, rv); ok && cmp == 0 {
				return joinRows(j.schema, j.curLeft, r), true, nil
			}
		}
		// Se acabó la derecha para esta izquierda: pasamos a la siguiente.
		j.haveLeft = false
	}
}

func (j *NestedLoopJoinOp) Close() error {
	_ = j.left.Close()
	return j.right.Close()
}

// ---------------------------------------------------------------------
// HashJoin
// ---------------------------------------------------------------------

type HashJoinOp struct {
	left, right           Operator
	leftTable, rightTable string
	leftCol, rightCol     string

	schema      catalog.Schema
	hash        map[string][]Row // derecha indexada por valor de la columna de join
	curMatches  []Row            // coincidencias pendientes de emitir para la izquierda actual
	curLeft     Row
	matchPos    int
	initialized bool
	leftIdx     int
}

// NewHashJoin crea un INNER JOIN por hash join.
func NewHashJoin(left Operator, leftTable, leftCol string, right Operator, rightTable, rightCol string) *HashJoinOp {
	return &HashJoinOp{
		left: left, right: right,
		leftTable: leftTable, rightTable: rightTable,
		leftCol: leftCol, rightCol: rightCol,
		schema: joinedSchema(left.Schema(), leftTable, right.Schema(), rightTable),
		hash:   make(map[string][]Row),
	}
}

func (j *HashJoinOp) Schema() catalog.Schema { return j.schema }

// build construye la tabla hash de la relación derecha (fase "build").
func (j *HashJoinOp) build() error {
	rightIdx := j.right.Schema().IndexOf(j.rightCol)
	for {
		row, ok, err := j.right.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		key := row.Values[rightIdx].String()
		j.hash[key] = append(j.hash[key], row)
	}
	j.leftIdx = j.left.Schema().IndexOf(j.leftCol)
	j.initialized = true
	return nil
}

func (j *HashJoinOp) Next() (Row, bool, error) {
	if !j.initialized {
		if err := j.build(); err != nil {
			return Row{}, false, err
		}
	}
	for {
		// Si hay coincidencias pendientes de la izquierda actual, las emitimos.
		if j.matchPos < len(j.curMatches) {
			r := j.curMatches[j.matchPos]
			j.matchPos++
			return joinRows(j.schema, j.curLeft, r), true, nil
		}
		// Si no, pedimos la siguiente fila izquierda y buscamos en el hash
		// (fase "probe").
		row, ok, err := j.left.Next()
		if err != nil || !ok {
			return Row{}, false, err
		}
		key := row.Values[j.leftIdx].String()
		j.curLeft = row
		j.curMatches = j.hash[key]
		j.matchPos = 0
	}
}

func (j *HashJoinOp) Close() error {
	_ = j.left.Close()
	return j.right.Close()
}
