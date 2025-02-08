package internal

var Defaults = &Inflector{}

type Inflector struct {
	rules map[RuleType]*Rule
}

func (i *Inflector) MustRegister(r *Rule) {
	if err := i.Register(r); err != nil {
		panic(err)
	}
}

func (i *Inflector) Register(r *Rule) error {
	if i.rules == nil {
		i.rules = make(map[RuleType]*Rule)
	}

	if err := r.Init(); err != nil {
		return err
	}

	i.rules[r.Type] = r

	return nil
}

func (i *Inflector) Inflected(tye RuleType, s string) string {
	if r, ok := i.rules[tye]; ok {
		return r.Inflected(s)
	}
	return s
}
