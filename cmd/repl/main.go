// REPL interactivo del motor de consultas SQL en memoria.
// Carga los CSV de testdata como tablas, luego lee consultas SELECT del
// usuario, las parsea (parser de Maicol), construye el árbol de operadores
// (exec de Daniel+Yokt) y muestra los resultados tabulados.
//
// Nota: por ahora solo conecta SELECT + WHERE (H2/H3). ORDER BY, LIMIT,
// GROUP BY (Victoria) y JOIN (Lennart) se integran aquí cuando estén listos.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/uss-taller-go/motor-sql-go/internal/catalog"
	"github.com/uss-taller-go/motor-sql-go/internal/exec"
	"github.com/uss-taller-go/motor-sql-go/internal/parser"
)

func main() {
	cat := catalog.NewCatalog()

	// Cargar las tablas de ejemplo disponibles en testdata.
	if tbl, err := catalog.LoadCSV("testdata/empleados.csv", "empleados"); err == nil {
		cat.Register(tbl)
	}

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

	// Construir el árbol de operadores: Scan -> Filter(WHERE) -> Project.
	var op exec.Operator = exec.NewScan(tbl)

	pred, err := exec.BuildPredicate(stmt.Where)
	if err != nil {
		return err
	}
	op = exec.NewFilter(op, pred)

	// SELECT * -> todas las columnas; si no, proyectar las pedidas.
	if !(len(stmt.Columns) == 1 && stmt.Columns[0] == "*") {
		op, err = exec.NewProject(op, stmt.Columns)
		if err != nil {
			return err
		}
	}
	defer op.Close()

	return printRows(op)
}

func printRows(op exec.Operator) error {
	schema := op.Schema()
	for _, col := range schema {
		fmt.Printf("%-15s", col.Name)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 15*len(schema)))

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
			fmt.Printf("%-15s", v.String())
		}
		fmt.Println()
		count++
	}
	fmt.Printf("(%d filas)\n\n", count)
	return nil
}
