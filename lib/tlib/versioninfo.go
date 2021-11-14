package tlib

// BuildTime is set at compile time
var BuildTime = ""

// BuildRef is set at compile time
var BuildRef = ""

// Ver returns version info
func Ver() string {
	return BuildTime + "_" + BuildRef
}
