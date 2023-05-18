package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"sync"
)

func evalBoolExpr(expression string, values map[string]bool) (bool, error) {
	// Parse the boolean expression and create the AST
	expr, err := parser.ParseExpr(expression)
	if err != nil {
		return false, fmt.Errorf("error parsing expression: %v", err)
	}

	// Create a custom visitor to walk the AST and evaluate the expression
	evalVisitor := &visitor{values: values}

	// Walk the AST and evaluate the expression
	ast.Walk(evalVisitor, expr)

	// Return the final result
	return evalVisitor.result, nil
}

type visitor struct {
	values map[string]bool // Input values for identifiers
	result bool            // Final result of the expression
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		// Skip nil nodes
		return v
	}

	switch expr := node.(type) {
	case *ast.Ident:
		// Check if the identifier exists in the input values
		value, ok := v.values[expr.Name]
		if !ok {
			panic(fmt.Errorf("identifier '%s' not found in input values", expr.Name))
		}
		v.result = value

	case *ast.UnaryExpr:
		// Handle unary expressions (e.g., !c)
		switch expr.Op {
		case token.NOT:
			childVisitor := &visitor{values: v.values}
			ast.Walk(childVisitor, expr.X)
			v.result = !childVisitor.result

		default:
			panic(fmt.Errorf("unsupported unary operator: %s", expr.Op))
		}

	case *ast.BinaryExpr:
		// Handle binary expressions (e.g., a && b)
		leftVisitor := &visitor{values: v.values}
		ast.Walk(leftVisitor, expr.X)
		rightVisitor := &visitor{values: v.values}
		ast.Walk(rightVisitor, expr.Y)

		switch expr.Op {
		case token.LAND:
			v.result = leftVisitor.result && rightVisitor.result
		case token.LOR:
			v.result = leftVisitor.result || rightVisitor.result
		default:
			panic(fmt.Errorf("unsupported binary operator: %s", expr.Op))
		}

	case *ast.ParenExpr:
		// Handle parentheses expressions
		childVisitor := &visitor{values: v.values}
		ast.Walk(childVisitor, expr.X)
		v.result = childVisitor.result

	default:
		panic(fmt.Errorf("unsupported expression type: %T", node))
	}

	return nil // Return nil to skip children nodes
}

func worker(i int, symbols []string, formula string, result chan map[string]bool, satisfied chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	// Compute the combination
	values := make(map[string]bool)
	for j, symbol := range symbols {
		value := (i>>j)&1 == 1
		values[symbol] = value
	}

	// Evaluate the expression on the computed combination of values
	res, err := evalBoolExpr(formula, values)
	if err != nil {
		panic(err)
	}

	// If the evaluation is true, return the result to the channel
	if res {
		satisfied <- true
        result <- values
	}
}

func main() {
	formulas := []string{
		"a && !a",
		"a || !a",
		"a && b || !c",
        "a && !b",
        "a && a",
	}
	symbols := []string{"a", "b", "c"}

	for _, formula := range formulas {
		// For each combination, eval the expression
		nCombinations := int(math.Pow(2, float64(len(symbols))))

		result := make(chan map[string]bool)
		satisfied := make(chan bool)
		var wg sync.WaitGroup

        // Launch worker threads
		for i := 0; i < nCombinations; i++ {
			wg.Add(1)
			go worker(i, symbols, formula, result, satisfied, &wg)
		}

        // Wait for the workers to finish in a goroutine
		go func() {
			wg.Wait()
			close(result)
			close(satisfied)
		}()

        // If the formula is satisfied print the result
		sat := <-satisfied
		if sat {
			resValues := <-result
            fmt.Printf("\033[97;1m%s\033[0m:\n", formula)
            fmt.Printf("  └─ \033[32msatisfied\033[0m by %v\n", resValues)
		} else {
            fmt.Printf("\033[97;1m%s\033[0m:\n", formula)
            fmt.Printf("  └─ \033[31munsatisfiable\033[0m\n")
		}
	}
}
