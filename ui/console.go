package ui

func Print(a ...interface{}) {
	printStyled(TypeLog, a...)
}

func Printf(format string, a ...interface{}) {
	printStyledf(TypeLog, format, a...)
}

func Println(a ...interface{}) {
	printStyledln(TypeLog, a...)
}

func Success(a ...interface{}) {
	printStyled(TypeSuccess, a...)
}

func Successf(format string, a ...interface{}) {
	printStyledf(TypeSuccess, format, a...)
}

func Error(a ...interface{}) {
	printStyled(TypeError, a...)
}

func Errorf(format string, a ...interface{}) {
	printStyledf(TypeError, format, a...)
}

func Warning(a ...interface{}) {
	printStyled(TypeWarning, a...)
}

func Warningf(format string, a ...interface{}) {
	printStyledf(TypeWarning, format, a...)
}

func Info(a ...interface{}) {
	printStyled(TypeInfo, a...)
}

func Infof(format string, a ...interface{}) {
	printStyledf(TypeInfo, format, a...)
}

func Debug(a ...interface{}) {
	printStyled(TypeDebug, a...)
}

func Debugf(format string, a ...interface{}) {
	printStyledf(TypeDebug, format, a...)
}
