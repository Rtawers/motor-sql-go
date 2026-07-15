package exec

import "github.com/uss-taller-go/motor-sql-go/internal/catalog"

type LimitOp struct {
	child Operator
	n     int
	count int
}

func NewLimit(child Operator, n int) *LimitOp {
	return &LimitOp{child: child, n: n}
}

func (l *LimitOp) Schema() catalog.Schema { return l.child.Schema() }

func (l *LimitOp) Next() (Row, bool, error) {
	if l.count >= l.n {
		return Row{}, false, nil
	}
	row, ok, err := l.child.Next()
	if err != nil || !ok {
		return Row{}, false, err
	}
	l.count++
	return row, true, nil
}

func (l *LimitOp) Close() error { return l.child.Close() }