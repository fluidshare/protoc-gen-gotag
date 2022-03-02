package module

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/fatih/structtag"
)

// Retag updates the existing tags with the map passed and modifies existing tags if any of the keys are matched.
// First key to the tags argument is the name of the struct, the second key corresponds to field names.
func Retag(n ast.Node, tags *StructTags) error {
	if n == nil {
		return nil
	}
	r := retag{}
	f := func(n ast.Node) ast.Visitor {
		if r.err != nil {
			return nil
		}

		if tp, ok := n.(*ast.TypeSpec); ok {
			if r.tags, ok = tags.Get(tp.Name.String()); ok {
				return r
			}
		}

		return nil
	}

	ast.Walk(structVisitor{f}, n)

	return r.err
}

type structVisitor struct {
	visitor func(n ast.Node) ast.Visitor
}

func (v structVisitor) Visit(n ast.Node) ast.Visitor {
	if tp, ok := n.(*ast.TypeSpec); ok {
		if _, ok = tp.Type.(*ast.StructType); ok {
			ast.Walk(v.visitor(n), n)
			return nil // This will ensure this struct is no longer traversed
		}
	}
	return v
}

type retag struct {
	err  error
	tags *FieldTags
}

func (v retag) Visit(n ast.Node) ast.Visitor {
	if v.err != nil {
		return nil
	}

	if f, ok := n.(*ast.Field); ok {
		if len(f.Names) == 0 {
			return nil
		}
		var newTags *structtag.Tags
		newTags, ok = v.tags.Get(f.Names[0].String())
		if !ok || newTags == nil {
			return nil
		}

		if f.Tag == nil {
			f.Tag = &ast.BasicLit{
				Kind: token.STRING,
			}
		}

		oldTags, err := structtag.Parse(strings.Trim(f.Tag.Value, "`"))
		if err != nil {
			v.err = err
			return nil
		}

		for _, t := range newTags.Tags() {
			oldTags.Set(t)
		}

		f.Tag.Value = "`" + oldTags.String() + "`"

		return nil
	}

	return v
}
