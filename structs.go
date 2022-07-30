package main

type Args struct {
	InPath  string `arg:"positional, required"`
	Mode    int    `arg:"-m" default:"1" help:"1 = random, 2 = growing."`
	OutPath string `arg:"-o, required" help:"Path to write output WebM to. Path will be made if it doesn't already exist."`
}
