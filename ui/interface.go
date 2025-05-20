package ui

type LogLevel uint8

const (
	TypeLog LogLevel = iota + 1
	TypeDebug
	TypeInfo
	TypeWarning
	TypeError
	TypeFatal
	TypeSuccess
)

type UI interface {
	Log(level LogLevel, msg string)
	Logf(level LogLevel, format string, args ...any)
	Progress() ProgressReporter
	Snippet() SnippetReporter
	Spinner() SpinnerReporter
	Table(headers []string, rows [][]string)
	Error(err error)
	Done(msg string)
}

type ProgressReporter interface {
	Start(total int, title string)
	Increment(msg string)
	Stop()
}

type LoaderReporter interface {
	Start(text string)
	Stop(text string)
}

type SnippetReporter interface {
	View(title string, data []byte)
}

type SpinnerReporter interface {
	Start(text string)
	Stop(text string)
}
