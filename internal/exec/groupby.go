package exec

import (
	"fmt"

	"github.com/uss-taller-go/motor-sql-go/internal/catalog"
	"github.com/uss-taller-go/motor-sql-go/internal/types"
)

type AggregateDef struct {
	Func   string // COUNT, SUM, AVG, MIN, MAX
	Column string
}

type GroupByOp struct {
	child    Operator
	groupCol string
	aggs     []AggregateDef
	schema   catalog.Schema
	results  []Row
	pos      int
	baked    bool
}

func NewGroupBy(child Operator, groupCol string, aggs []AggregateDef) (*GroupByOp, error) {
	var sch catalog.Schema
	childSch := child.Schema()

	if groupCol != "" {
		idx := childSch.IndexOf(groupCol)
		if idx < 0 {
			return nil, fmt.Errorf("exec: columna %q no existe para GROUP BY", groupCol)
		}
		sch = append(sch, childSch[idx])
	}

	for _, agg := range aggs {
		kind := types.KindFloat
		if agg.Func == "COUNT" {
			kind = types.KindInt
		} else if agg.Column != "*" {
			idx := childSch.IndexOf(agg.Column)
			if idx >= 0 {
				kind = childSch[idx].Kind
				if agg.Func == "AVG" {
					kind = types.KindFloat
				}
			}
		}
		name := fmt.Sprintf("%s(%s)", agg.Func, agg.Column)
		sch = append(sch, catalog.Column{Name: name, Kind: kind})
	}

	return &GroupByOp{child: child, groupCol: groupCol, aggs: aggs, schema: sch}, nil
}

func (g *GroupByOp) Schema() catalog.Schema { return g.schema }

func (g *GroupByOp) Next() (Row, bool, error) {
	if !g.baked {
		if err := g.bake(); err != nil {
			return Row{}, false, err
		}
	}
	if g.pos >= len(g.results) {
		return Row{}, false, nil
	}
	row := g.results[g.pos]
	g.pos++
	return row, true, nil
}

func (g *GroupByOp) bake() error {
	groups := make(map[string][]Row)
	var groupKeys []string // Para iterar en orden de llegada

	groupColIdx := -1
	if g.groupCol != "" {
		groupColIdx = g.child.Schema().IndexOf(g.groupCol)
	}

	for {
		row, ok, err := g.child.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}

		key := "ALL"
		if groupColIdx >= 0 {
			key = row.Values[groupColIdx].String()
		}

		if _, exists := groups[key]; !exists {
			groupKeys = append(groupKeys, key)
		}
		groups[key] = append(groups[key], row)
	}

	for _, key := range groupKeys {
		rows := groups[key]
		outValues := make([]types.Value, 0, len(g.schema))

		if groupColIdx >= 0 {
			outValues = append(outValues, rows[0].Values[groupColIdx])
		}

		for _, agg := range g.aggs {
			val, err := computeAgg(agg, rows)
			if err != nil {
				return err
			}
			outValues = append(outValues, val)
		}

		g.results = append(g.results, Row{Schema: g.schema, Values: outValues})
	}

	g.baked = true
	return nil
}

func computeAgg(agg AggregateDef, rows []Row) (types.Value, error) {
	if agg.Func == "COUNT" {
		if agg.Column == "*" {
			return types.Int(int64(len(rows))), nil
		}
		colIdx := rows[0].Schema.IndexOf(agg.Column)
		if colIdx < 0 {
			return types.Null, fmt.Errorf("columna %q no existe para agregación", agg.Column)
		}
		count := int64(0)
		for _, r := range rows {
			if !r.Values[colIdx].IsNull() {
				count++
			}
		}
		return types.Int(count), nil
	}

	colIdx := rows[0].Schema.IndexOf(agg.Column)
	if colIdx < 0 {
		return types.Null, fmt.Errorf("columna %q no existe para agregación", agg.Column)
	}

	var sum float64
	var hasVals bool
	var minVal, maxVal types.Value

	for _, r := range rows {
		v := r.Values[colIdx]
		if v.IsNull() {
			continue // Convención SQL: ignorar NULLs en agregaciones
		}
		hasVals = true

		if agg.Func == "MIN" || agg.Func == "MAX" {
			if minVal.IsNull() {
				minVal = v
				maxVal = v
			} else {
				cmp, ok := types.Compare(v, minVal)
				if ok && cmp < 0 {
					minVal = v
				}
				cmpMax, ok := types.Compare(v, maxVal)
				if ok && cmpMax > 0 {
					maxVal = v
				}
			}
		}

		if agg.Func == "SUM" || agg.Func == "AVG" {
			if v.Kind == types.KindInt {
				sum += float64(v.I)
			} else if v.Kind == types.KindFloat {
				sum += v.F
			}
		}
	}

	if !hasVals && agg.Func != "COUNT" {
		return types.Null, nil
	}

	switch agg.Func {
	case "SUM":
		if rows[0].Schema[colIdx].Kind == types.KindInt {
			return types.Int(int64(sum)), nil
		}
		return types.Float(sum), nil
	case "AVG":
		count := 0
		for _, r := range rows {
			if !r.Values[colIdx].IsNull() {
				count++
			}
		}
		return types.Float(sum / float64(count)), nil
	case "MIN":
		return minVal, nil
	case "MAX":
		return maxVal, nil
	}

	return types.Null, fmt.Errorf("agregado %q no soportado", agg.Func)
}

func (g *GroupByOp) Close() error { return g.child.Close() }
