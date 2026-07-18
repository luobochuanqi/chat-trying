package skills

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

func executeCalculator(args map[string]interface{}) (string, error) {
	expression, ok := args["expression"].(string)
	if !ok || expression == "" {
		return "", fmt.Errorf("missing expression parameter")
	}
	result, err := evalExpression(expression)
	if err != nil {
		return fmt.Sprintf("计算错误: %s", err.Error()), nil
	}
	return strconv.FormatFloat(result, 'f', -1, 64), nil
}

func evalExpression(expr string) (float64, error) {
	expr = strings.NewReplacer("×", "*", "÷", "/", "x", "*", "X", "*", "（", "(", "）", ")").Replace(expr)
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, err
	}
	return evalNode(node)
}

func evalNode(node ast.Expr) (float64, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		if n.Kind == token.INT || n.Kind == token.FLOAT {
			return strconv.ParseFloat(n.Value, 64)
		}
		return 0, fmt.Errorf("unsupported literal: %s", n.Value)
	case *ast.BinaryExpr:
		left, err := evalNode(n.X)
		if err != nil {
			return 0, err
		}
		right, err := evalNode(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		case token.REM:
			return 0, fmt.Errorf("modulo not supported")
		default:
			return 0, fmt.Errorf("unsupported operator: %s", n.Op)
		}
	case *ast.ParenExpr:
		return evalNode(n.X)
	case *ast.UnaryExpr:
		val, err := evalNode(n.X)
		if err != nil {
			return 0, err
		}
		if n.Op == token.SUB {
			return -val, nil
		}
		return val, nil
	}
	return 0, fmt.Errorf("unsupported expression type")
}
