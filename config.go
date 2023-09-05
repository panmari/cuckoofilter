package cuckoo

type FilterPrecision uint

const (
	Medium FilterPrecision = iota
	Low
	High
)

type Config struct {
	NumElements uint
	Precision   FilterPrecision
}
