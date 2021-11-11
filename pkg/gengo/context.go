package gengo

import (
	"context"
	"fmt"
	"go/types"
	"sort"
	"strings"

	"github.com/go-courier/logr"

	gengotypes "github.com/octohelm/gengo/pkg/types"
	"github.com/pkg/errors"
)

func NewContext(args *GeneratorArgs) (*Context, error) {
	u, pkgPaths, err := gengotypes.Load(args.Entrypoint)
	if err != nil {
		return nil, err
	}
	c := &Context{
		PkgPaths: pkgPaths,
		Args:     args,
		Universe: u,
	}
	return c, nil
}

type Context struct {
	PkgPaths map[string]bool
	Args     *GeneratorArgs
	Universe gengotypes.Universe
	Package  gengotypes.Package
	Tags     map[string][]string
}

func (c *Context) Execute(ctx context.Context, generators ...Generator) error {
	for pkgPath := range c.PkgPaths {
		if err := c.pkgExecute(ctx, pkgPath, generators...); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) pkgExecute(ctx context.Context, pkg string, generators ...Generator) error {
	p, ok := c.Universe[pkg]
	if !ok {
		return errors.Errorf("invalid pkg `%s`", pkg)
	}

	pkgCtx := &Context{
		Args:     c.Args,
		Package:  p,
		Universe: c.Universe,
		Tags:     map[string][]string{},
	}

	for _, f := range p.Files() {
		if f.Doc != nil && len(f.Doc.List) > 0 {
			tags, _ := gengotypes.ExtractCommentTags(strings.Split(f.Doc.Text(), "\n"))
			for k := range tags {
				pkgCtx.Tags[k] = tags[k]
			}
		}
	}

	gfs := genfiles{}

	for i := range generators {
		g, err := generators[i].Init(pkgCtx, gfs)
		if err != nil {
			return err
		}

		if e := pkgCtx.doGenerate(ctx, g); e != nil {
			return errors.Wrapf(e, "`%s` generate failed for %s", g.Name(), pkgCtx.Package.Pkg().Path())
		}
	}

	for _, w := range gfs {
		if err := w.WriteToFile(pkgCtx); err != nil {
			return err
		}
	}

	return nil
}

func (c *Context) doGenerate(ctx context.Context, g Generator) error {
	if c.Package == nil {
		return nil
	}

	pkgTypes := c.Package.Types()

	names := make([]string, 0)

	for n := range pkgTypes {
		names = append(names, n)
	}

	sort.Strings(names)

	for _, n := range names {
		tpe := pkgTypes[n].Type()

		// skip type XXX interface{}
		if _, ok := tpe.Underlying().(*types.Interface); ok {
			continue
		}

		if named, ok := tpe.(*types.Named); ok {
			// 	skip alias other pkg type XXX = XXX2
			if named.Obj().Pkg() != c.Package.Pkg() {
				continue
			}

			tags, _ := c.Universe.Package(named.Obj().Pkg().Path()).Doc(named.Obj().Pos())

			mergedTags := merge(c.Tags, tags)

			if isGeneratorEnabled(g, mergedTags) {
				shouldProcess := true

				if typeFilter, ok := g.(GeneratorTypeFilter); ok {
					shouldProcess = typeFilter.FilterType(c, named)

				}

				if shouldProcess {
					if err := g.GenerateType(c, named); err != nil {
						return err
					}
					logr.FromContext(ctx).Debug(fmt.Sprintf("`%s` gen found for  %s.", g.Name(), named))
				}
			}
		}
	}

	return nil
}

func isGeneratorEnabled(g Generator, tags map[string][]string) bool {
	values, ok := tags["gengo:"+g.Name()]
	return ok && strings.Join(values, "") != "false"
}
