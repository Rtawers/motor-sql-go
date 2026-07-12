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

## H1 — Interfaz Operator: Next() con (Row, bool, error) vs. sentinel io.EOF — 2026-07-08 — Daniel

**Qué se decidió:** (pendiente — el equipo debe discutirlo y llenarlo)

**Qué otras opciones se evaluaron y por qué se descartaron:**

**Razón técnica de la elección:**

**¿Se consultó IA?:** Sí — se usó Claude para armar el esqueleto inicial del
repositorio (estructura de paquetes, interfaz Operator de referencia,
carga de CSV, tests table-driven de ejemplo) como punto de partida del
equipo. El diseño final de cada módulo, su justificación y la implementación
de los hitos restantes son responsabilidad de cada integrante — deben poder
explicar y modificar en vivo cualquier parte del código propio en la defensa
oral.

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