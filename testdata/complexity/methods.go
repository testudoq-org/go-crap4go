// Package methods contains method declarations to verify method naming
// and CC extraction for receiver types including pointer receivers.
package methods

// Greeter holds a greeting prefix.
type Greeter struct {
	Prefix string
}

// Greet is a value-receiver method: named "Greeter.Greet", CC = 1.
func (g Greeter) Greet(name string) string {
	return g.Prefix + name
}

// SetPrefix is a pointer-receiver method: named "Greeter.SetPrefix", CC = 1.
// The pointer receiver "*Greeter" is normalised to "Greeter" in the display name.
func (g *Greeter) SetPrefix(p string) {
	g.Prefix = p
}

// ConditionalGreet has an if statement: named "Greeter.ConditionalGreet", CC = 2.
func (g *Greeter) ConditionalGreet(name string) string {
	if name == "" {
		return g.Prefix + "stranger"
	}
	return g.Prefix + name
}

// StandaloneFunc is a package-level function (not a method): named "StandaloneFunc", CC = 1.
func StandaloneFunc() string {
	return "hello"
}
