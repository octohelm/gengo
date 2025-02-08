package internal

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
)

type RuleType int

const (
	Plural RuleType = iota
	Singular
)

type RuleItem struct {
	Pattern     string
	Replacement string
}

type IrregularItem struct {
	Word        string
	Replacement string
}

type CompiledRule struct {
	Replacement string
	Regexp      *regexp.Regexp
}

type Rule struct {
	Type      RuleType
	Rules     []*RuleItem
	Irregular []*IrregularItem

	uninflected         []string
	compiledIrregular   *regexp.Regexp
	compiledUninflected *regexp.Regexp
	compiledRules       []*CompiledRule

	irregularMap map[string]string

	cache sync.Map
}

func (r *Rule) Inflected(s string) string {
	inflected, _ := r.cache.LoadOrStore(s, sync.OnceValue(func() string {
		return r.inflected(s)
	}))
	return inflected.(func() string)()
}

func (r *Rule) inflected(s string) string {
	if res := r.compiledIrregular.FindStringSubmatch(s); len(res) >= 3 {
		var buf strings.Builder

		buf.WriteString(res[1])
		buf.WriteString(s[0:1])
		buf.WriteString(r.irregularMap[strings.ToLower(res[2])][1:])

		return buf.String()
	}

	if r.compiledUninflected.MatchString(s) {
		return s
	}

	for _, re := range r.compiledRules {
		if re.Regexp.MatchString(s) {
			return re.Regexp.ReplaceAllString(s, re.Replacement)
		}
	}

	return s
}

func (r *Rule) Init() error {
	var reString string

	switch r.Type {
	case Plural:
		r.uninflected = slices.Concat(uninflected, uninflectedPlurals)
	case Singular:
		r.uninflected = slices.Concat(uninflected, uninflectedSingulars)
	}

	reString = fmt.Sprintf(`(?i)(^(?:%s))$`, strings.Join(r.uninflected, `|`))

	r.compiledUninflected = regexp.MustCompile(reString)

	r.irregularMap = make(map[string]string, len(r.Irregular))

	vIrregulars := make([]string, len(r.Irregular))
	for i, item := range r.Irregular {
		vIrregulars[i] = item.Word
		r.irregularMap[item.Word] = item.Replacement
	}

	reString = fmt.Sprintf(`(?i)(.*)\b((?:%s))$`, strings.Join(vIrregulars, `|`))
	r.compiledIrregular = regexp.MustCompile(reString)

	r.compiledRules = make([]*CompiledRule, len(r.Rules))
	for i, item := range r.Rules {
		r.compiledRules[i] = &CompiledRule{item.Replacement, regexp.MustCompile(item.Pattern)}
	}

	return nil
}

var (
	uninflected = []string{
		`Amoyese`, `bison`, `Borghese`, `bream`, `breeches`, `britches`, `buffalo`,
		`cantus`, `carp`, `chassis`, `clippers`, `cod`, `coitus`, `Congoese`,
		`contretemps`, `corps`, `debris`, `diabetes`, `djinn`, `eland`, `elk`,
		`equipment`, `Faroese`, `flounder`, `Foochowese`, `gallows`, `Genevese`,
		`Genoese`, `Gilbertese`, `graffiti`, `headquarters`, `herpes`, `hijinks`,
		`Hottentotese`, `information`, `innings`, `jackanapes`, `Kiplingese`,
		`Kongoese`, `Lucchese`, `mackerel`, `Maltese`, `.*?media`, `mews`, `moose`,
		`mumps`, `Nankingese`, `news`, `nexus`, `Niasese`, `Pekingese`,
		`Piedmontese`, `pincers`, `Pistoiese`, `pliers`, `Portuguese`, `proceedings`,
		`rabies`, `rice`, `rhinoceros`, `salmon`, `Sarawakese`, `scissors`,
		`sea[- ]bass`, `series`, `Shavese`, `shears`, `siemens`, `species`, `swine`,
		`testes`, `trousers`, `trout`, `tuna`, `Vermontese`, `Wenchowese`, `whiting`,
		`wildebeest`, `Yengeese`,
	}
	uninflectedPlurals = []string{
		`.*[nrlm]ese`, `.*deer`, `.*fish`, `.*measles`, `.*ois`, `.*pox`, `.*sheep`,
		`people`,
	}

	uninflectedSingulars = []string{
		`.*[nrlm]ese`, `.*deer`, `.*fish`, `.*measles`, `.*ois`, `.*pox`, `.*sheep`,
		`.*ss`,
	}
)
