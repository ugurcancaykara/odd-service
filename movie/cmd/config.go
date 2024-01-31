package main

type serviceConfig struct {
	API apiConfig `yaml:"api"`
}

type apiConfig struct {
	Port int `yaml:"port"`
}
