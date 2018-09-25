package expression

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

func ParseAndEval(exp string) (bool, error) {
	exp = strings.ToLower(exp)
	exp = strings.Replace(exp, "and", "&&", -1)
	exp = strings.Replace(exp, "or", "||", -1)
	tree, err := parser.ParseExpr(exp)
	if err != nil {
		return false, err
	}
	return eval(tree)
}

func eval(tree ast.Expr) (bool, error) {
	switch n := tree.(type) {
	case *ast.Ident:
		v := strings.ToLower(n.String())
		if v != "true" && v != "false" {
			return unsup(v)
		}
		return v == "true", nil
	case *ast.BinaryExpr:
		if n.Op != token.LOR && n.Op != token.LAND {
			return unsup(n.Op)
		}
		x, err := eval(n.X)
		if err != nil {
			return false, err
		}
		y, err := eval(n.Y)
		if err != nil {
			return false, err
		}
		if n.Op == token.LAND {
			return x && y, nil
		}
		return x || y, nil
	case *ast.ParenExpr:
		return eval(n.X)
	}
	return unsup(reflect.TypeOf(tree))
}

func unsup(i interface{}) (bool, error) {
	return false, errors.New(fmt.Sprintf("%v unsupported", i))
}
