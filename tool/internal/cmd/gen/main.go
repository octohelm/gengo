package main

import (
	"context"
	"flag"

	"github.com/octohelm/x/logr"
	"github.com/octohelm/x/logr/slog"

	"github.com/octohelm/gengo/pkg/gengo"

	_ "github.com/octohelm/gengo/devpkg/deepcopygen"

	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
)

func main() {
	flag.Parse()

	c, err := gengo.NewExecutor(&gengo.GeneratorArgs{
		Entrypoint:         flag.Args(),
		OutputFileBaseName: "zz_generated",
	})
	if err != nil {
		panic(err)
	}

	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	if err := c.Execute(ctx, gengo.GetRegisteredGenerators()...); err != nil {
		panic(err)
	}
}
