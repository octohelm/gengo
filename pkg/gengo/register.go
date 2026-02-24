package gengo

var registeredGenerators = map[string]Generator{}

// GetRegisteredGenerators 返回全部已注册生成器，或在给出 names 时返回指定子集。
func GetRegisteredGenerators(names ...string) (generators []Generator) {
	if len(names) == 0 {
		for name := range registeredGenerators {
			generators = append(generators, registeredGenerators[name])
		}
		return generators
	}

	for _, name := range names {
		if _, ok := registeredGenerators[name]; ok {
			generators = append(generators, registeredGenerators[name])
		}
	}

	return generators
}

// Register 按名称把 g 注册到全局生成器表。
func Register(g Generator) {
	registeredGenerators[g.Name()] = g
}
