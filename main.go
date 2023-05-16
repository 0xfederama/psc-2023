package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

func main() {
	{
		expression := "a && b || !c"
		values := map[string]bool{
			"a": true,
			"b": false,
			"c": true,
		}

		result, err := evalBoolExpr(expression, values)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Result: %5v, with values %v\n", result, values)
		}
	}
	{
		expression := "a || !a"
		values := map[string]bool{
			"a": true,
		}

		result, err := evalBoolExpr(expression, values)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Result: %5v, with values %v\n", result, values)
		}
	}
}
