# Motor de consultas SQL en memoria

Proyecto del Taller de Programación en Go — grupo: Maicol, Pamela, Victoria, Lennart y Daniel (encargado).

Motor que parsea un subconjunto de SQL y lo ejecuta sobre tablas cargadas
desde CSV en memoria, usando un modelo de ejecución por operadores
iteradores (modelo Volcano).

## Requisitos

- Go 1.22+

## Compilar y ejecutar

```bash
go build ./...
go run ./cmd/repl
```

## Tests

```bash
go test ./...
go test -race ./...      # obligatorio una vez haya concurrencia
```

## Estructura del repositorio

```
cmd/repl/          CLI/REPL — punto de entrada del programa
internal/types/    Sistema de tipos de valores (Value, comparación, NULL)
internal/catalog/  Carga de CSV, esquema, catálogo de tablas          [Pamela]
internal/lexer/    Tokenizador de SQL                                 [Maicol]
internal/parser/   Parser de descenso recursivo -> AST                [Maicol]
internal/ast/      Definición de nodos del AST                        [Maicol]
internal/exec/      Interfaz Operator + Scan/Filter/Project (núcleo)  [Daniel]
                    OrderBy/Limit/GroupBy + agregados                 [Victoria]
                    Join (nested-loop y hash join)                    [Lennart]
testdata/          CSVs de ejemplo para pruebas y demos
docs/              Gramática soportada, bitácora de decisiones, declaración de IA
```

## Estado (hitos)

- [x] H1 — Carga de CSV a catálogo en memoria (`internal/catalog`)
- [x] Núcleo de ejecución: `Operator`, `Scan`, `Filter`, `Project` (base para H3)
- [ ] H2 — Lexer + parser SQL -> AST
- [ ] H3 — CLI/REPL leyendo consultas reales del usuario
- [ ] H4 — ORDER BY, LIMIT, GROUP BY + agregados, manejo de NULL
- [ ] H5 — INNER JOIN (nested-loop y hash join)

## Convención de commits

Commits pequeños y descriptivos, prefijados por módulo/hito, por ejemplo:

```
catalog: inferencia de tipos desde primera fila de datos
exec: agregar operador Filter con evaluación perezosa
parser: soporte para WHERE con AND/OR y paréntesis
```

Evitar commits masivos que junten varios hitos o módulos.
