// REPL interactivo del motor de consultas SQL en memoria.
//
// Integra el trabajo de todo el equipo: carga tablas (catalog/Pamela),
// parsea la consulta (parser/Maicol) y construye el árbol de operadores
// (exec) aplicándolos en el orden correcto de un motor SQL:
//
//	Scan -> Join -> Filter(WHERE) -> GroupBy -> OrderBy -> Limit -> Project
//
// PROPIEDAD: Daniel + Yokt (integración final).
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/uss-taller-go/motor-sql-go/internal/ast"
	"github.com/uss-taller-go/motor-sql-go/internal/catalog"
	"github.com/uss-taller-go/motor-sql-go/internal/exec"
	"github.com/uss-taller-go/motor-sql-go/internal/parser"
)

func main() {
	cat := catalog.NewCatalog()

	// Cargar las tablas de ejemplo disponibles en testdata.
	loadAllCSV(cat, "testdata")

	fmt.Println("Motor SQL en memoria — REPL")
	fmt.Println("Tablas cargadas:", strings.Join(cat.TableNames(), ", "))
	fmt.Println("Escribe una consulta SELECT (o 'salir' para terminar).")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("sql> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "salir" || line == "exit" || line == "quit" {
			break
		}
		if err := runQuery(cat, line); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error de lectura:", err)
	}
}

func loadInto(cat *catalog.Catalog, path, name string) {
	if tbl, err := catalog.LoadCSV(path, name); err == nil {
		cat.Register(tbl)
	}
}

// loadAllCSV carga todos los archivos .csv del directorio dado en el catálogo.
func loadAllCSV(cat *catalog.Catalog, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".csv") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".csv")
		loadInto(cat, filepath.Join(dir, e.Name()), name)
	}
}

// runQuery ejecuta una consulta: parse -> plan -> ejecución -> impresión.
func runQuery(cat *catalog.Catalog, sql string) error {
	stmt, err := parser.Parse(sql)
	if err != nil {
		return err
	}

	tbl, ok := cat.Table(stmt.From)
	if !ok {
		return fmt.Errorf("la tabla %q no existe", stmt.From)
	}

	// 1. Scan de la tabla base.
	var op exec.Operator = exec.NewScan(tbl)

	// 2. JOIN (si hay). Combina la tabla base con otra del catálogo.
	if stmt.Join != nil {
		op, err = buildJoin(cat, op, stmt.From, stmt.Join)
		if err != nil {
			return err
		}
	}

	// 3. Filter (WHERE). Tras un JOIN, resolvemos los nombres de columna del
	// WHERE contra el esquema real (que tiene nombres prefijados tabla.col).
	where := stmt.Where
	if stmt.Join != nil {
		where = resolveExprColumns(where, op.Schema())
	}
	pred, err := exec.BuildPredicate(where)
	if err != nil {
		return err
	}
	op = exec.NewFilter(op, pred)

	// 4. GroupBy + agregados (si hay GROUP BY o algún agregado en el SELECT).
	groupCol, aggs, hasAgg := extractAggregates(stmt)
	if len(stmt.GroupBy) > 0 || hasAgg {
		if len(stmt.GroupBy) > 0 {
			groupCol = stmt.GroupBy[0] // el motor agrupa por una columna
		}
		op, err = exec.NewGroupBy(op, groupCol, aggs)
		if err != nil {
			return err
		}
	}

	// 5. OrderBy (si hay). El motor ordena por la primera columna indicada.
	if len(stmt.OrderBy) > 0 {
		first := stmt.OrderBy[0]
		op = exec.NewOrderBy(op, first.Column, first.Desc)
	}

	// 6. Limit (si hay).
	if stmt.Limit != nil {
		op = exec.NewLimit(op, *stmt.Limit)
	}

	// 7. Project (SELECT de columnas simples). Si hay agregados, el GroupBy
	// ya produjo el esquema final, así que no proyectamos.
	if !hasAgg && len(stmt.GroupBy) == 0 {
		cols := simpleColumns(stmt.SelectItems)
		if !(len(cols) == 1 && cols[0] == "*") {
			// Tras un JOIN las columnas quedan prefijadas (tabla.col); si el
			// usuario las pidió sin prefijo, las resolvemos contra el esquema.
			cols = resolveColumns(op.Schema(), cols)
			op, err = exec.NewProject(op, cols)
			if err != nil {
				return err
			}
		}
	}

	defer op.Close()
	return printRows(op)
}

// buildJoin arma el operador de join a partir de la cláusula ON, que es una
// expresión del tipo tablaA.colA = tablaB.colB.
func buildJoin(cat *catalog.Catalog, left exec.Operator, leftTable string, jc *ast.JoinClause) (exec.Operator, error) {
	rightTbl, ok := cat.Table(jc.Table)
	if !ok {
		return nil, fmt.Errorf("la tabla %q del JOIN no existe", jc.Table)
	}
	right := exec.NewScan(rightTbl)

	leftCol, rightCol, err := joinColumns(jc.On, leftTable, jc.Table)
	if err != nil {
		return nil, err
	}
	// Usamos hash join por defecto (más eficiente).
	return exec.NewHashJoin(left, leftTable, leftCol, right, jc.Table, rightCol), nil
}

// joinColumns extrae las columnas de la condición ON (tablaA.colA = tablaB.colB).
func joinColumns(on ast.Expr, leftTable, rightTable string) (leftCol, rightCol string, err error) {
	be, ok := on.(*ast.BinaryExpr)
	if !ok || be.Op != "=" {
		return "", "", fmt.Errorf("la condición del JOIN debe ser una igualdad col = col")
	}
	l, lok := be.Left.(*ast.ColumnRef)
	r, rok := be.Right.(*ast.ColumnRef)
	if !lok || !rok {
		return "", "", fmt.Errorf("la condición del JOIN debe comparar dos columnas")
	}
	lname := stripTable(l.Name)
	rname := stripTable(r.Name)
	// Asignar cada columna a su tabla según el prefijo, si lo tiene.
	if strings.HasPrefix(r.Name, rightTable+".") || strings.HasPrefix(l.Name, leftTable+".") {
		return lname, rname, nil
	}
	return lname, rname, nil
}

// stripTable quita el prefijo "tabla." de un nombre de columna si lo tiene.
func stripTable(name string) string {
	if i := strings.Index(name, "."); i >= 0 {
		return name[i+1:]
	}
	return name
}

// extractAggregates recorre el SELECT y devuelve los agregados encontrados.
func extractAggregates(stmt *ast.SelectStmt) (groupCol string, aggs []exec.AggregateDef, hasAgg bool) {
	for _, item := range stmt.SelectItems {
		if a, ok := item.(*ast.AggregateItem); ok {
			aggs = append(aggs, exec.AggregateDef{Func: a.Func, Column: a.Column})
			hasAgg = true
		}
	}
	return groupCol, aggs, hasAgg
}

// simpleColumns extrae los nombres de las columnas simples del SELECT.
func simpleColumns(items []ast.SelectItem) []string {
	var cols []string
	for _, item := range items {
		if c, ok := item.(*ast.ColumnItem); ok {
			cols = append(cols, c.Name)
		}
	}
	if len(cols) == 0 {
		return []string{"*"}
	}
	return cols
}

// resolveColumns empareja los nombres pedidos por el usuario con los del
// esquema real: acepta "nombre" aunque el esquema lo tenga como "tabla.nombre".
// resolveExprColumns recorre una expresión y reescribe cada ColumnRef sin
// prefijo por su nombre prefijado del esquema (tabla.col), tras un JOIN.
func resolveExprColumns(e ast.Expr, schema catalog.Schema) ast.Expr {
	switch node := e.(type) {
	case *ast.BinaryExpr:
		return &ast.BinaryExpr{
			Left:  resolveExprColumns(node.Left, schema),
			Op:    node.Op,
			Right: resolveExprColumns(node.Right, schema),
		}
	case *ast.ColumnRef:
		if schema.IndexOf(node.Name) >= 0 {
			return node
		}
		for _, col := range schema {
			if stripTable(col.Name) == node.Name {
				return &ast.ColumnRef{Name: col.Name}
			}
		}
		return node
	default:
		return e
	}
}

func resolveColumns(schema catalog.Schema, wanted []string) []string {
	out := make([]string, len(wanted))
	for i, w := range wanted {
		out[i] = w
		if schema.IndexOf(w) >= 0 {
			continue // coincidencia exacta
		}
		for _, col := range schema {
			if stripTable(col.Name) == w {
				out[i] = col.Name
				break
			}
		}
	}
	return out
}

func printRows(op exec.Operator) error {
	schema := op.Schema()
	for _, col := range schema {
		fmt.Printf("%-18s", col.Name)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 18*len(schema)))

	count := 0
	for {
		row, ok, err := op.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		for _, v := range row.Values {
			fmt.Printf("%-18s", v.String())
		}
		fmt.Println()
		count++
	}
	fmt.Printf("(%d filas)\n\n", count)
	return nil
}
