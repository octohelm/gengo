package gengo

import "context"

type Project struct {
	Entrypoint []string `arg:""`
	NoCache    bool     `flag:",omitzero"`
}

func (g *Project) Run(ctx context.Context) error {
	c, err := NewExecutor(&GeneratorArgs{
		Entrypoint:         g.Entrypoint,
		NoCache:            g.NoCache,
		OutputFileBaseName: "zz_generated",
	})
	if err != nil {
		return err
	}
	return c.Execute(ctx, GetRegisteredGenerators()...)
}
