// Package types define el sistema de tipos de valores que circulan por el
// motor: enteros, decimales, texto, booleanos y NULL. Todos los paquetes
// (catalog, exec, lexer/parser) dependen de este paquete, nunca al revés.
package types

import "fmt"

// Kind identifica el tipo dinámico de un Value.
type Kind int

const (
	KindNull Kind = iota
	KindInt
	KindFloat
	KindString
	KindBool
)

func (k Kind) String() string {
	switch k {
	case KindNull:
		return "NULL"
	case KindInt:
		return "INT"
	case KindFloat:
		return "FLOAT"
	case KindString:
		return "STRING"
	case KindBool:
		return "BOOL"
	default:
		return "UNKNOWN"
	}
}

// Value representa un valor de columna con su tipo dinámico.
// Se pasa por valor (no puntero): son inmutables y pequeños.
type Value struct {
	Kind Kind
	I    int64
	F    float64
	S    string
	B    bool
}

// Null es el valor NULL compartido.
var Null = Value{Kind: KindNull}

func Int(v int64) Value     { return Value{Kind: KindInt, I: v} }
func Float(v float64) Value { return Value{Kind: KindFloat, F: v} }
func Str(v string) Value    { return Value{Kind: KindString, S: v} }
func Bool(v bool) Value     { return Value{Kind: KindBool, B: v} }

func (v Value) IsNull() bool { return v.Kind == KindNull }

func (v Value) String() string {
	switch v.Kind {
	case KindNull:
		return "NULL"
	case KindInt:
		return fmt.Sprintf("%d", v.I)
	case KindFloat:
		return fmt.Sprintf("%g", v.F)
	case KindString:
		return v.S
	case KindBool:
		return fmt.Sprintf("%t", v.B)
	default:
		return "?"
	}
}

// Compare compara dos valores del mismo tipo (o numéricamente compatibles).
// Devuelve (-1, 0, 1) igual que strings.Compare, y ok=false si la comparación
// no es válida (tipos incompatibles) o si alguno es NULL (NULL no compara).
//
// DECISIÓN DE DISEÑO PENDIENTE PARA EL EQUIPO:
// ¿Cómo se comportan las comparaciones con NULL en WHERE? El estándar SQL
// dice que `NULL = NULL` es UNKNOWN (no true). Documenten esta decisión en
// la bitácora (Hito H4) — no la tomen por default sin discutirla.
func Compare(a, b Value) (cmp int, ok bool) {
	if a.IsNull() || b.IsNull() {
		return 0, false
	}
	switch {
	case a.Kind == KindInt && b.Kind == KindInt:
		return compareInt64(a.I, b.I), true
	case a.Kind == KindFloat || b.Kind == KindFloat:
		af, aok := asFloat(a)
		bf, bok := asFloat(b)
		if !aok || !bok {
			return 0, false
		}
		return compareFloat64(af, bf), true
	case a.Kind == KindString && b.Kind == KindString:
		if a.S == b.S {
			return 0, true
		}
		if a.S < b.S {
			return -1, true
		}
		return 1, true
	case a.Kind == KindBool && b.Kind == KindBool:
		if a.B == b.B {
			return 0, true
		}
		if !a.B && b.B {
			return -1, true
		}
		return 1, true
	default:
		return 0, false
	}
}

func asFloat(v Value) (float64, bool) {
	switch v.Kind {
	case KindFloat:
		return v.F, true
	case KindInt:
		return float64(v.I), true
	default:
		return 0, false
	}
}

func compareInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareFloat64(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
