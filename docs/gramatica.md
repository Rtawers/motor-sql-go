# Gramática SQL soportada (EBNF)

Documento vivo — actualizar a medida que el parser (Maicol) avance por los
hitos. Esta es la meta H2-H5; el parser puede implementarse incrementalmente.

```ebnf
select_stmt   ::= "SELECT" select_list "FROM" table_ref
                   [ "WHERE" expr ]
                   [ "GROUP BY" column_list ]
                   [ "ORDER BY" order_list ]
                   [ "LIMIT" integer ] ;

select_list   ::= "*" | select_item { "," select_item } ;
select_item   ::= (column | aggregate_call) [ "AS" identifier ] ;
aggregate_call::= ("COUNT" | "SUM" | "AVG" | "MIN" | "MAX") "(" ("*" | column) ")" ;

table_ref     ::= identifier [ "AS" identifier ]
                 [ join_clause ] ;
join_clause   ::= "INNER" "JOIN" identifier [ "AS" identifier ] "ON" expr ;

column_list   ::= column { "," column } ;
order_list    ::= order_item { "," order_item } ;
order_item    ::= column [ "ASC" | "DESC" ] ;

expr          ::= or_expr ;
or_expr       ::= and_expr { "OR" and_expr } ;
and_expr      ::= comparison { "AND" comparison } ;
comparison    ::= operand comp_op operand | "(" expr ")" ;
comp_op       ::= "=" | "<>" | "<" | ">" | "<=" | ">=" ;
operand       ::= column | literal ;

column        ::= identifier [ "." identifier ] ;
literal       ::= integer | float | string | "true" | "false" | "NULL" ;
identifier    ::= letter { letter | digit | "_" } ;
```

## Notas de implementación (para la bitácora de Maicol)

- Decidir: ¿descenso recursivo o Pratt parsing para `expr`? Pratt es más
  natural para precedencia de `AND`/`OR`/comparaciones y para escalar a
  aritmética en el futuro.
- Los errores de sintaxis deben incluir mensaje y posición (línea/columna o
  al menos offset de token) — requisito explícito de la rúbrica (20%).
- El AST debe vivir en `internal/ast`, separado del parser, para que
  `internal/exec` pueda consumirlo sin depender del lexer/parser.
