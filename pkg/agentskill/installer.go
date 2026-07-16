package agentskill

import "context"

type SkillsInstaller struct{}

func (g *SkillsInstaller) Run(ctx context.Context) error {
	return (&Installer{}).Install(ctx)
}
