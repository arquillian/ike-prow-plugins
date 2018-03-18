package log

// Logger is a facade interface compatible with logrus.Logger but with limited scope of logging.
// It exists to decouple plugin implementations from particular log implementation but also to only allow
// reporting actionable problems and corner cases such as panic (or in other words to avoid logging as program flow analysis / debugging).
type Logger interface {
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
}
