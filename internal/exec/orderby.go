package exec

import (
	"fmt"
	"sort"

	"github.com/uss-taller-go/motor-sql-go/internal/catalog"
	"github.com/uss-taller-go/motor-sql-go/internal/types"
)

type OrderByOp struct {
	child    Operator
	col      string
	desc     bool
	sorted   []Row
	pos      int
	baked    bool
	colIndex int
}

func NewOrderBy(child Operator, col string, desc bool) *OrderByOp {
	return &OrderByOp{child: child, col: col, desc: desc, colIndex: -1}
}

func (o *OrderByOp) Schema() catalog.Schema { return o.child.Schema() }

func (o *OrderByOp) Next() (Row, bool, error) {
	if !o.baked {
		if err := o.bake(); err != nil {
			return Row{}, false, err
		}
	}
	if o.pos >= len(o.sorted) {
		return Row{}, false, nil
	}
	row := o.sorted[o.pos]
	o.pos++
	return row, true, nil
}

func (o *OrderByOp) bake() error {
	o.colIndex = o.child.Schema().IndexOf(o.col)
	// MEJORA: Validar si la columna existe (pedido en la guía de H4)
	if o.colIndex < 0 {
		return fmt.Errorf("exec: la columna %q no existe en el esquema para ORDER BY", o.col)
	}

	for {
		row, ok, err := o.child.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		o.sorted = append(o.sorted, row)
	}

	sort.SliceStable(o.sorted, func(i, j int) bool {
		vi := o.sorted[i].Values[o.colIndex]
		vj := o.sorted[j].Values[o.colIndex]
		cmp, ok := types.Compare(vi, vj)
		if !ok {
			return false // NULL u orden indefinido se mantienen estables
		}
		if o.desc {
			return cmp > 0
		}
		return cmp < 0
	})

	o.baked = true
	return nil
}

func (o *OrderByOp) Close() error { return o.child.Close() }
