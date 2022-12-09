package ghttpex

// Logger use in Client to log something
type Logger interface {
	Println(v ...interface{})
}
