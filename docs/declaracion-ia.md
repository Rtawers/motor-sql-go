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
