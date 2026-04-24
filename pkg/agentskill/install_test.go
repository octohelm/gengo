package agentskill

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"
)

func TestParseGoModSkills(t *testing.T) {
	t.Parallel()

	skills := MustValue(t, func() ([]SkillRef, error) {
		return ParseGoModSkills([]byte(`
module github.com/example/project

require (
	// +skill:first
	github.com/example/one v1.2.3
	// +skill:second
	// +skill:third
	github.com/example/two v1.2.4 // indirect
)
`))
	})

	Then(t, "解析出三个 skill 引用",
		Expect(skills,
			Be(cmp.Len[[]SkillRef](3)),
		),
		Expect(skills, Equal([]SkillRef{
			{Name: "first", Module: "github.com/example/one", Version: "v1.2.3"},
			{Name: "second", Module: "github.com/example/two", Version: "v1.2.4"},
			{Name: "third", Module: "github.com/example/two", Version: "v1.2.4"},
		})),
	)
}

func TestPlanSkillInstall(t *testing.T) {
	t.Parallel()

	t.Run("WHEN 模块与 skill 都存在", func(t *testing.T) {
		root := t.TempDir()
		modCache := t.TempDir()

		firstSource := mustCreateCachedSkill(t, modCache, SkillRef{
			Name:    "first",
			Module:  "github.com/example/one",
			Version: "v1.2.3",
		})
		secondSource := mustCreateCachedSkill(t, modCache, SkillRef{
			Name:    "second",
			Module:  "github.com/example/two",
			Version: "v1.2.4",
		})

		plan := MustValue(t, func() (*InstallPlan, error) {
			return PlanSkillInstall(root, modCache, []SkillRef{
				{Name: "first", Module: "github.com/example/one", Version: "v1.2.3"},
				{Name: "second", Module: "github.com/example/two", Version: "v1.2.4"},
			})
		})

		Then(t, "生成按目标目录组织的安装计划",
			Expect(plan.GitIgnoreNames, Equal([]string{"first", "second"})),
			Expect(plan.Skills, Equal([]SkillInstall{
				{
					Ref:         SkillRef{Name: "first", Module: "github.com/example/one", Version: "v1.2.3"},
					ModuleRoot:  filepath.Join(modCache, "github.com/example/one@v1.2.3"),
					Source:      firstSource,
					Destination: filepath.Join(root, ".agents", "skills", "first"),
				},
				{
					Ref:         SkillRef{Name: "second", Module: "github.com/example/two", Version: "v1.2.4"},
					ModuleRoot:  filepath.Join(modCache, "github.com/example/two@v1.2.4"),
					Source:      secondSource,
					Destination: filepath.Join(root, ".agents", "skills", "second"),
				},
			})),
		)
	})

	t.Run("WHEN 模块版本不在 cache", func(t *testing.T) {
		root := t.TempDir()
		modCache := t.TempDir()

		Then(t, "返回明确的模块 cache 缺失错误",
			ExpectDo(func() error {
				_, err := PlanSkillInstall(root, modCache, []SkillRef{
					{Name: "first", Module: "github.com/example/one", Version: "v1.2.3"},
				})
				return err
			},
				ErrorMatch(regexp.MustCompile(`module "github\.com/example/one"@v1\.2\.3 not found in module cache`)),
			),
		)
	})
}

func TestApplyInstallPlan(t *testing.T) {
	t.Parallel()

	t.Run("WHEN 执行安装计划", func(t *testing.T) {
		root := t.TempDir()
		modCache := t.TempDir()

		firstSource := mustCreateCachedSkill(t, modCache, SkillRef{
			Name:    "first",
			Module:  "github.com/example/one",
			Version: "v1.2.3",
		})
		secondSource := mustCreateCachedSkill(t, modCache, SkillRef{
			Name:    "second",
			Module:  "github.com/example/two",
			Version: "v1.2.4",
		})

		plan := MustValue(t, func() (*InstallPlan, error) {
			return PlanSkillInstall(root, modCache, []SkillRef{
				{Name: "first", Module: "github.com/example/one", Version: "v1.2.3"},
				{Name: "second", Module: "github.com/example/two", Version: "v1.2.4"},
			})
		})

		Must(t, func() error {
			return ApplyInstallPlan(plan)
		})

		Then(t, "把 skill 目录软链到项目内并写入 gitignore",
			Expect(mustEvalSymlink(t, filepath.Join(root, ".agents", "skills", "first")), Equal(mustEvalSymlink(t, firstSource))),
			Expect(mustEvalSymlink(t, filepath.Join(root, ".agents", "skills", "second")), Equal(mustEvalSymlink(t, secondSource))),
			Expect(mustReadFile(t, filepath.Join(root, ".agents", "skills", ".gitignore")), Equal("first\nsecond\n")),
		)
	})

	t.Run("WHEN 目标路径已存在", func(t *testing.T) {
		root := t.TempDir()
		modCache := t.TempDir()

		source := mustCreateCachedSkill(t, modCache, SkillRef{
			Name:    "first",
			Module:  "github.com/example/one",
			Version: "v1.2.3",
		})

		dst := filepath.Join(root, ".agents", "skills", "first")
		Must(t, func() error { return os.MkdirAll(dst, 0o755) })
		Must(t, func() error { return os.WriteFile(filepath.Join(dst, "stale.txt"), []byte("stale"), 0o644) })

		plan := MustValue(t, func() (*InstallPlan, error) {
			return PlanSkillInstall(root, modCache, []SkillRef{
				{Name: "first", Module: "github.com/example/one", Version: "v1.2.3"},
			})
		})

		Must(t, func() error {
			return ApplyInstallPlan(plan)
		})

		info := MustValue(t, func() (os.FileInfo, error) {
			return os.Lstat(dst)
		})

		Then(t, "强制覆盖现有路径并重建软链",
			Expect(info.Mode()&os.ModeSymlink != 0, Equal(true)),
			Expect(mustEvalSymlink(t, dst), Equal(mustEvalSymlink(t, source))),
		)
	})
}

func TestInstallerPlanAndInstall(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	modCache := t.TempDir()

	Must(t, func() error {
		return os.WriteFile(filepath.Join(root, "go.mod"), []byte(`module github.com/example/project

require (
	// +skill:first
	github.com/example/one v1.2.3
)
`), 0o644)
	})

	source := mustCreateCachedSkill(t, modCache, SkillRef{
		Name:    "first",
		Module:  "github.com/example/one",
		Version: "v1.2.3",
	})

	installer := &Installer{
		ProjectRoot: root,
		GoModCache:  modCache,
	}

	plan := MustValue(t, func() (*InstallPlan, error) {
		return installer.Plan(context.Background())
	})

	Must(t, func() error {
		return installer.Install(context.Background())
	})

	Then(t, "Installer 通过计划驱动安装",
		Expect(plan.Skills, Be(cmp.Len[[]SkillInstall](1))),
		Expect(mustEvalSymlink(t, filepath.Join(root, ".agents", "skills", "first")), Equal(mustEvalSymlink(t, source))),
	)
}

func mustCreateCachedSkill(t *testing.T, modCache string, skill SkillRef) string {
	t.Helper()

	_, dir, err := cachedSkillPath(modCache, skill)
	if err != nil {
		t.Fatal(err)
	}

	Must(t, func() error { return os.MkdirAll(dir, 0o755) })
	Must(t, func() error { return os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+skill.Name), 0o644) })

	return dir
}

func mustEvalSymlink(t *testing.T, path string) string {
	t.Helper()

	return MustValue(t, func() (string, error) {
		return filepath.EvalSymlinks(path)
	})
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()

	return string(MustValue(t, func() ([]byte, error) {
		return os.ReadFile(path)
	}))
}
