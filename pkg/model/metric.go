package model

// Nep17Balance describes universal balance task configuration.
type Nep17Balance struct {
	Contract    string   `yaml:"contract"`
	TotalSupply bool     `yaml:"totalSupply"`
	BalanceOf   []string `yaml:"balanceOf"`
}
