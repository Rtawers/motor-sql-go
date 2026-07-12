# Bitácora de decisiones

Una entrada por cada decisión de diseño relevante. No busca extensión, sino
evidencia de razonamiento propio (Anexo A del documento del curso).

Plantilla por entrada:

```
## [Hito] — [Título corto de la decisión] — [Fecha] — [Autor]

**Qué se decidió:**

**Qué otras opciones se evaluaron y por qué se descartaron:**

**Razón técnica de la elección:**

**¿Se consultó IA? ¿Qué se preguntó y qué se hizo con la respuesta
(aceptar/adaptar/descartar)?**
```

---

## H1 — Interfaz Operator: Next() con (Row, bool, error) vs. sentinel io.EOF — 2026-07-12 — Daniel + Yokt

**Qué se decidió:** El método Next() de la interfaz Operator devuelve
(Row, bool, error). El bool (ok) en false indica fin de datos (fin normal);
el error se reserva solo para fallos reales. Fin de datos y error nunca se
confunden.

**Qué otras opciones se evaluaron y por qué se descartaron:** Se evaluó la
convención idiomática de Go de señalar el fin con un error centinela io.EOF
(como hace io.Reader en la librería estándar), devolviendo solo (Row, error).
Se descartó porque obliga a quien consume la interfaz a recordar chequear
io.EOF antes que cualquier otro error; si un integrante lo olvida, trata el
fin normal como un error y rompe la ejecución.

**Razón técnica de la elección:** Somos seis personas y varios operadores
(Filter, Project, y luego los de Victoria y Lennart) consumen esta interfaz.
El bool explícito es imposible de malinterpretar y hace la integración entre
módulos más segura. Priorizamos la claridad de integración del equipo sobre
la elegancia idiomática, que pesa más en código de librería para terceros
que en un motor interno.

**¿Se consultó IA?:** Sí — se usó Claude para armar el esqueleto inicial del
repositorio (estructura de paquetes, interfaz Operator de referencia, carga
de CSV, tests table-driven de ejemplo). La decisión sobre la convención de
Next() la tomó el equipo tras comparar ambas opciones; se aceptó la que ya
venía implementada por las razones anteriores.
---

## H1 — Inferencia de tipos en catalog.InferType — 2026-11-07 — Pamela

**Qué se decidió:** 
Se implementó una inferencia de tipos automática escaneando hacia abajo columna por columna. Si la primera celda está vacía, el iterador sigue bajando hasta encontrar el primer valor válido y deduce el tipo (`KindInt`, `KindFloat`, `KindBool`, o `KindString`). Además, se configuró `csv.Reader` con `FieldsPerRecord = -1` para tolerar filas con columnas faltantes y tratarlas como `NULL`.

**Qué otras opciones se evaluaron y por qué se descartaron:** 
Se evaluó exigir un archivo externo separado (tipo `.schema`) para definir los tipos explícitamente, pero se descartó para mantener la facilidad de uso del motor en memoria. También se evaluó dejar que fallara ante columnas incompletas, pero descartamos el comportamiento restrictivo para facilitar las pruebas con datos "sucios".

**Razón técnica de la elección:** 
La librería estándar de Go lee todo como `string` y puede paniquear si las comas no coinciden. Necesitábamos un tipado dinámico pero estricto y tolerante a fallos para que la función `types.Compare` no rompa el motor cuando los operadores `Filter` y `Project` procesen datos nulos.

**¿Se consultó IA?:** 
Sí, se utilizó Gemini para revisar la lógica del `TODO` en `catalog.go` y adaptar correctamente el lector `csv.Reader` a las exigencias de casos nulos/faltantes.

---

(agregar una entrada nueva por cada decisión relevante de H2 en adelante)


## H2 — Estrategia de Parsing: Descenso Recursivo vs Pratt — [Fecha actual] — Maicol

**Qué se decidió:** 
Se implementó un parser por descenso recursivo (Recursive Descent Parsing) para analizar la gramática SQL, evaluar las expresiones del `WHERE` y construir el AST. La precedencia de operadores se resolvió anidando llamadas a funciones (`parseOrExpr` llama a `parseAndExpr`, que llama a `parseComparison`).

**Qué otras opciones se evaluaron y por qué se descartaron:** 
Se evaluó usar el algoritmo de Pratt (Top-Down Operator Precedence). Aunque Pratt es más escalable para expresiones aritméticas complejas, se descartó porque el subconjunto SQL actual (solo condicionales lógicos `AND`/`OR` y comparaciones) es manejable con descenso recursivo. 

**Razón técnica de la elección:** 
El descenso recursivo mapea de forma directa y literal uno a uno con la gramática EBNF definida en el proyecto. Esto hace que el código de `parser.go` sea mucho más fácil de leer, seguir y depurar (especialmente útil para lanzar errores exactos de línea y columna al fallar una regla).

**¿Se consultó IA? ¿Qué se preguntó y qué se hizo con la respuesta?** 
Sí, se utilizó Gemini para estructurar la base del `Lexer`, generar las funciones del `Parser` por descenso recursivo, definir las interfaces del AST y construir la suite de pruebas *table-driven* de `parser_test.go`. También se usó para identificar y corregir un bug donde el lexer no estaba tokenizando el carácter asterisco (`*`).

---

## H3 — Semántica de NULL en el WHERE — 2026-07-12 — Daniel + Yokt

**Qué se decidió:** Cualquier comparación que involucre un NULL (ej.
salario > 3000 con salario NULL) se evalúa a NULL, no a true ni false. Como
el WHERE solo deja pasar filas cuyo resultado sea booleano true, una fila con
NULL en la comparación se descarta.

**Qué otras opciones se evaluaron y por qué se descartaron:** Se evaluó tratar
NULL como "menor que todo" o como cero, para que las comparaciones siempre
den true o false. Se descartó porque contradice el estándar SQL y produce
resultados engañosos (un salario desconocido no es "menor que 3000").

**Razón técnica de la elección:** Sigue la semántica de tres valores del SQL
real (true/false/unknown). La función types.Compare ya devuelve ok=false
cuando hay un NULL, así que el traductor solo tuvo que interpretar ese caso
como "la fila no pasa".

**¿Se consultó IA?:** Sí — se usó Claude para diseñar el traductor AST a
Predicate y decidir cómo propagar el NULL. El equipo revisó y validó la
semántica contra el estándar SQL antes de aceptarla.


## H3 — Puente AST a Predicate entre parser y motor de ejecución — 2026-07-12 — Daniel + Yokt

**Qué se decidió:** Se implementó BuildPredicate en internal/exec, que recorre
recursivamente la expresión del WHERE (ast.Expr de Maicol) y produce un
Predicate que el operador Filter evalúa fila por fila. Separa operadores
lógicos (AND/OR) de comparaciones (=, <>, <, >, <=, >=).

**Qué otras opciones se evaluaron y por qué se descartaron:** Se evaluó que el
parser produjera directamente el Predicate, pero se descartó para no mezclar
responsabilidades: el parser solo entiende sintaxis (produce el AST), y el
motor de ejecución es el único que sabe cómo evaluar una fila. Esto respeta
la separación entre análisis sintáctico y ejecución.

**Razón técnica de la elección:** Mantener el traductor en exec permite que el
Filter siga siendo genérico (recibe cualquier Predicate) y que agregar nuevos
operadores del AST no obligue a tocar el parser. Es coherente con el principio
abierto/cerrado que exige la rúbrica.

**¿Se consultó IA?:** Sí — se usó Claude para escribir el traductor y su suite
de tests de integración (parser -> traductor -> filtro). El equipo comprende
el flujo completo y puede modificarlo en vivo.