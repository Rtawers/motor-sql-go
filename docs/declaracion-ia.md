# Declaración de uso de IA

Conforme al Anexo B del documento del curso. Declarar honestamente no
penaliza; ocultarlo sí.

## Entrega inicial — esqueleto del repositorio

**Nombre de los asistentes/herramientas de IA empleados:** Claude (Anthropic).

**Para qué se usó:** Generación del esqueleto inicial del repositorio:
estructura de carpetas/paquetes Go, definición de referencia de la interfaz
`Operator` (modelo Volcano) con implementaciones de `Scan`, `Filter` y
`Project`, un paquete `types` básico de valores, carga de CSV en `catalog`,
tests table-driven de ejemplo, y documentación de soporte (README, esta
plantilla, plantilla de bitácora).

**En qué módulos o archivos influyó:** `internal/types/value.go`,
`internal/catalog/catalog.go`, `internal/exec/operator.go`,
`internal/exec/operator_test.go`, `cmd/repl/main.go`, `README.md`,
`docs/bitacora.md`, `docs/gramatica.md` (versión inicial).

**Qué partes son de autoría íntegra del equipo:** El lexer y parser SQL
(`internal/lexer`, `internal/parser`, `internal/ast`), la traducción de AST a
árbol de operadores, los operadores de H4 (OrderBy/Limit/GroupBy y
agregados) y H5 (Join), el CLI/REPL interactivo real, y toda decisión de
diseño documentada en la bitácora — incluyendo cualquier modificación al
esqueleto inicial generado.

> «Declaro que soy autor del diseño y la lógica central de este proyecto,
> que comprendo todo el código entregado y que puedo explicarlo y
> modificarlo.»
>
> — Firma / nombre: _____________________ (cada integrante debe firmar por
> su módulo antes de la entrega final)

---

(agregar una entrada nueva por cada uso adicional de IA durante el proyecto,
por hito o por integrante)


## Hito 2 (H2) — Lexer, Parser y AST

**Nombre de los asistentes/herramientas de IA empleados:** Gemini.

**Para qué se usó:** 
Implementación completa del analizador léxico (`Lexer`), el analizador sintáctico por descenso recursivo (`Parser`), la definición de nodos del Árbol de Sintaxis Abstracta (`AST`), y la creación de la suite de pruebas basadas en tablas (table-driven tests) para validar consultas exitosas y capturar errores de sintaxis con posición exacta (línea/columna).

**En qué módulos o archivos influyó:** 
`internal/lexer/lexer.go`, `internal/parser/parser.go`, `internal/parser/parser_test.go`, y `internal/ast/ast.go`.

**Qué partes son de autoría íntegra del equipo:** 
La revisión, asimilación del funcionamiento del descenso recursivo, la ejecución de pruebas locales para certificar los requisitos de la rúbrica y las correcciones iterativas de bugs detectados durante el testing (ej. soporte para el token `*`).

> «Declaro que soy autor del diseño y la lógica central de este proyecto,
> que comprendo todo el código entregado y que puedo explicarlo y
> modificarlo.»
>
> — Firma / nombre: Maicol Rafael

## Hito 4 (H4) — Ordenamiento y Agregaciones

**Nombre de los asistentes/herramientas de IA empleados:** Gemini.

**Para qué se usó:** 
Diseño y desarrollo de los operadores relacionales `LimitOp` (streaming) y `GroupByOp` (materializador). Generación de la lógica de evaluación de agregados y el agrupamiento dinámico en memoria (`bake`), asegurando el filtrado correcto de valores nulos, junto con las pruebas de validación automatizadas correspondientes.

**En qué módulos o archivos influyó:** 
`internal/exec/orderby.go`, `internal/exec/limit.go`, `internal/exec/groupby.go` y todos sus archivos de testing asociados (`_test.go`).

**Qué partes son de autoría íntegra del equipo:** 
La decisión sobre el flujo del pipeline (streaming vs materialización en memoria), el manejo estricto de tipos de salida y las convenciones adoptadas para la evaluación de NULLs documentadas en la bitácora.

> «Declaro que soy autor del diseño y la lógica central de este proyecto,
> que comprendo todo el código entregado y que puedo explicarlo y
> modificarlo.»
>
> — Firma / nombre: Lucrecia Victoria