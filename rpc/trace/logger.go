package trace

const (
	NotPrintVerbose = 0
	ErrPrintVerbose = 1
	AllPrintVerbose = 2
)

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

type Logger struct {
	verbose int
}

func (p *Logger) NewTrace(verbose int, traceID, clientName, serviceMethod string) Trace {
	return nil
}
