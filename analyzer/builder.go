package analyzer

func NewConfig(l Language) *Config {
	return &Config{
		Tokenizer: AlphaNumericTokenizer,
		TokenFilters: []TokenFilter{
			LowercaseFilter,
			SynonymFilter, // skipped if SynonymFilterMapping empty
			NormalizeFilter,
			CompoundFilter, // language based CompoundFilter used if CompoundFilterDictionary empty
			StemFilter,
			StopWordFilter,
			UniqueFilter,
			LengthFilter,
		},
		Options: Options{
			Language: l,
		},
	}
}

func (c *Config) WithLanguage(l Language) *Config {
	c.Options.Language = l
	return c
}

func (c *Config) WithSynonyms(m [][]string) *Config {
	c.Options.SynonymFilterMapping = m
	return c
}

func (c *Config) WithCompound(d []string) *Config {
	c.Options.CompoundFilterDictionary = d
	return c
}

func (c *Config) WithTokenizer(t Tokenizer) *Config {
	c.Tokenizer = t
	return c
}

func (c *Config) WithFilter(f TokenFilter) *Config {
	c.TokenFilters = append(c.TokenFilters, f)
	return c
}

func (c *Config) WithLength(min, max int) *Config {
	c.Options.LengthFilterMin = min
	c.Options.LengthFilterMax = max
	return c
}

func (c *Config) WithoutStem() *Config {
	for i, f := range c.TokenFilters {
		if f == StemFilter {
			c.TokenFilters = append(c.TokenFilters[:i], c.TokenFilters[i+1:]...)
			break
		}
	}
	return c
}
