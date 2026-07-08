// Package lexer tokeniza texto SQL en internal/lexer -> internal/parser.
//
// PROPIEDAD: Maicol.
// Esqueleto de partida: define el tipo Token y sus tipos, sin implementar
// aún el escaneo real. Ver docs/gramatica.md para el subconjunto de SQL a
// cubrir.
package lexer

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenInt
	TokenFloat
	TokenString
	TokenKeyword  // SELECT, FROM, WHERE, AND, OR, ORDER, BY, GROUP, LIMIT, ASC, DESC, JOIN, ON, INNER, AS...
	TokenOperator // =, <>, <, >, <=, >=, ,, (, )
	TokenIllegal  // token no reconocido -> el parser lo traduce a error con posición
)

// Token es la unidad mínima producida por el lexer.
type Token struct {
	Type TokenType
	Lit  string // literal/lexema tal cual apareció en el texto
	Pos  int    // offset en el texto fuente, para mensajes de error con posición
	Line int
	Col  int
}

// TODO(maicol, H2): implementar Lexer con un método Next() Token (o similar)
// que recorra el string de entrada carácter a carácter, reconociendo
// identificadores, keywords, números, strings entre comillas, operadores y
// puntuación, y reportando TokenIllegal con posición ante caracteres no
// válidos.
