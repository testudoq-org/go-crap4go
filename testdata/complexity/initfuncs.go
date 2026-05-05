// Package initfuncs contains multiple init() functions to verify that the
// complexity extractor disambiguates them with unique names.
package initfuncs

var order []string

func init() {
	order = append(order, "first")
}

func init() {
	order = append(order, "second")
}

func init() {
	order = append(order, "third")
}
