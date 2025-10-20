package types

type Config struct {
	Stores map[string]LocalStore `yaml:"stores"`
}
