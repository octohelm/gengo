package agentskill

import (
	"bytes"
	"context"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/module"
)

type SkillRef struct {
	Name    string
	Module  string
	Version string
}

type SkillInstall struct {
	Ref         SkillRef
	ModuleRoot  string
	Source      string
	Destination string
}

type InstallPlan struct {
	SkillsDir      string
	GitIgnorePath  string
	GitIgnoreNames []string
	Skills         []SkillInstall
}

type Installer struct {
	ProjectRoot string
	GoModPath   string
	GoModCache  string
}

func (i *Installer) Install(ctx context.Context) error {
	plan, err := i.Plan(ctx)
	if err != nil {
		return err
	}

	return ApplyInstallPlan(plan)
}

func (i *Installer) Plan(ctx context.Context) (*InstallPlan, error) {
	projectRoot, err := i.projectRoot(ctx)
	if err != nil {
		return nil, err
	}

	goModPath := i.GoModPath
	if goModPath == "" {
		goModPath = filepath.Join(projectRoot, "go.mod")
	}

	skills, err := ParseGoModSkillsFile(goModPath)
	if err != nil {
		return nil, err
	}

	goModCache, err := i.goModCache(ctx)
	if err != nil {
		return nil, err
	}

	return PlanSkillInstall(projectRoot, goModCache, skills)
}

func PlanSkillInstall(projectRoot string, goModCache string, skills []SkillRef) (*InstallPlan, error) {
	skillsDir := filepath.Join(projectRoot, ".agents", "skills")

	plan := &InstallPlan{
		SkillsDir:      skillsDir,
		GitIgnorePath:  filepath.Join(skillsDir, ".gitignore"),
		GitIgnoreNames: make([]string, 0, len(skills)),
		Skills:         make([]SkillInstall, 0, len(skills)),
	}

	for _, skill := range skills {
		install, err := planSkillInstall(skillsDir, goModCache, skill)
		if err != nil {
			return nil, err
		}

		plan.GitIgnoreNames = append(plan.GitIgnoreNames, skill.Name)
		plan.Skills = append(plan.Skills, install)
	}

	return plan, nil
}

func ApplyInstallPlan(plan *InstallPlan) error {
	if err := os.MkdirAll(plan.SkillsDir, 0o755); err != nil {
		return fmt.Errorf("create skills dir: %w", err)
	}

	for _, skill := range plan.Skills {
		if err := ensureSymlink(skill.Destination, skill.Source); err != nil {
			return fmt.Errorf("install skill %q: %w", skill.Ref.Name, err)
		}
	}

	if err := ensureGitIgnore(plan.GitIgnorePath, plan.GitIgnoreNames); err != nil {
		return err
	}

	return nil
}

func planSkillInstall(skillsDir string, goModCache string, skill SkillRef) (SkillInstall, error) {
	moduleRoot, source, err := cachedSkillPath(goModCache, skill)
	if err != nil {
		return SkillInstall{}, err
	}

	moduleInfo, err := os.Stat(moduleRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return SkillInstall{}, fmt.Errorf("module %q@%s not found in module cache: %s", skill.Module, skill.Version, moduleRoot)
		}
		return SkillInstall{}, fmt.Errorf("stat module cache %q: %w", moduleRoot, err)
	}
	if !moduleInfo.IsDir() {
		return SkillInstall{}, fmt.Errorf("module cache path %q is not a directory", moduleRoot)
	}

	info, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return SkillInstall{}, fmt.Errorf("skill %q not found under module cache path: %s", skill.Name, source)
		}
		return SkillInstall{}, fmt.Errorf("stat skill source %q: %w", source, err)
	}
	if !info.IsDir() {
		return SkillInstall{}, fmt.Errorf("skill source %q is not a directory", source)
	}

	return SkillInstall{
		Ref:         skill,
		ModuleRoot:  moduleRoot,
		Source:      source,
		Destination: filepath.Join(skillsDir, skill.Name),
	}, nil
}

func ParseGoModSkillsFile(path string) ([]SkillRef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}

	skills, err := ParseGoModSkills(data)
	if err != nil {
		return nil, fmt.Errorf("parse go.mod skills: %w", err)
	}

	return skills, nil
}

func ParseGoModSkills(data []byte) ([]SkillRef, error) {
	var out []SkillRef

	inRequireBlock := false
	var pending []string

	for lineNo, raw := range bytes.Split(data, []byte("\n")) {
		line := strings.TrimSpace(string(raw))

		switch {
		case line == "":
			pending = nil
			continue
		case line == "require (":
			inRequireBlock = true
			pending = nil
			continue
		case inRequireBlock && line == ")":
			inRequireBlock = false
			pending = nil
			continue
		}

		if strings.HasPrefix(line, "//") {
			if skill, ok := parseSkillComment(line); ok {
				pending = append(pending, skill)
			} else {
				pending = nil
			}
			continue
		}

		modulePath, version, ok := parseRequireLine(line, inRequireBlock)
		if !ok {
			pending = nil
			continue
		}

		for _, name := range pending {
			out = append(out, SkillRef{
				Name:    name,
				Module:  modulePath,
				Version: version,
			})
		}
		pending = nil

		if modulePath == "" || version == "" {
			return nil, fmt.Errorf("invalid require line %d", lineNo+1)
		}
	}

	return out, nil
}

func parseSkillComment(line string) (string, bool) {
	const prefix = "// +skill:"

	if !strings.HasPrefix(line, prefix) {
		return "", false
	}

	name := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if name == "" {
		return "", false
	}

	return name, true
}

func parseRequireLine(line string, inRequireBlock bool) (string, string, bool) {
	if strings.HasPrefix(line, "exclude ") || strings.HasPrefix(line, "replace ") || strings.HasPrefix(line, "retract ") {
		return "", "", false
	}

	if !inRequireBlock {
		if !strings.HasPrefix(line, "require ") {
			return "", "", false
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "require "))
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", "", false
	}

	return fields[0], fields[1], true
}

func cachedSkillPath(goModCache string, skill SkillRef) (string, string, error) {
	escapedModule, err := module.EscapePath(skill.Module)
	if err != nil {
		return "", "", fmt.Errorf("escape module path %q: %w", skill.Module, err)
	}

	escapedVersion, err := module.EscapeVersion(skill.Version)
	if err != nil {
		return "", "", fmt.Errorf("escape module version %q: %w", skill.Version, err)
	}

	moduleRoot := filepath.Join(goModCache, escapedModule+"@"+escapedVersion)
	return moduleRoot, filepath.Join(moduleRoot, ".agents", "skills", skill.Name), nil
}

func ensureSymlink(dst string, src string) error {
	if info, err := os.Lstat(dst); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(dst); err != nil {
				return fmt.Errorf("remove existing symlink %q: %w", dst, err)
			}
		} else {
			if err := os.RemoveAll(dst); err != nil {
				return fmt.Errorf("remove existing path %q: %w", dst, err)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat destination %q: %w", dst, err)
	}

	if err := os.Symlink(src, dst); err != nil {
		return fmt.Errorf("create symlink %q -> %q: %w", dst, src, err)
	}

	return nil
}

func ensureGitIgnore(path string, names []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create gitignore dir: %w", err)
	}

	existing := map[string]struct{}{}
	lines := make([]string, 0, len(names))

	if data, err := os.ReadFile(path); err == nil {
		for line := range strings.SplitSeq(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			existing[trimmed] = struct{}{}
			lines = append(lines, trimmed)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read gitignore %q: %w", path, err)
	}

	unique := make([]string, 0, len(names))
	seen := map[string]struct{}{}
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		unique = append(unique, name)
	}
	sort.Strings(unique)

	for _, name := range unique {
		if _, ok := existing[name]; ok {
			continue
		}
		lines = append(lines, name)
	}

	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write gitignore %q: %w", path, err)
	}

	return nil
}

func (i *Installer) projectRoot(ctx context.Context) (string, error) {
	_ = ctx

	if i.ProjectRoot != "" {
		return i.ProjectRoot, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for dir := cwd; ; dir = filepath.Dir(dir) {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("stat %q: %w", goModPath, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	return "", fmt.Errorf("go.mod not found from current working directory")
}

func (i *Installer) goModCache(ctx context.Context) (string, error) {
	_ = ctx

	if i.GoModCache != "" {
		return i.GoModCache, nil
	}

	if goModCache := strings.TrimSpace(os.Getenv("GOMODCACHE")); goModCache != "" {
		return goModCache, nil
	}

	gopath := build.Default.GOPATH
	if gopath == "" {
		return "", fmt.Errorf("cannot resolve module cache: GOPATH is empty")
	}

	paths := filepath.SplitList(gopath)
	if len(paths) == 0 || paths[0] == "" {
		return "", fmt.Errorf("cannot resolve module cache from GOPATH: %q", gopath)
	}

	return filepath.Join(paths[0], "pkg", "mod"), nil
}
