package trace

type Logger struct {
	out     Output
	verbose int
}

func NewLogger() *Logger {
	return &Logger{
		out:     DefaultOutput,
		verbose: 0,
	}
}

func (p *Logger) SetOutput(out Output) {
	p.out = out
}

func (p *Logger) GetVerbose() int {
	return p.verbose
}

func (p *Logger) SetVerbose(verbose int) {
	p.verbose = verbose
}

func (p *Logger) NewTrace(server bool, verbose int, traceID, clientName, clientAddr, serverName, serverAddr, classMethod string) Trace {
	out := p.out
	if out == nil {
		return nopTrace{}
	}

	v := max(p.verbose, verbose)
	if v < 0 {
		return nopTrace{}
	} else if v == 0 {
		return &errTrace{
			out:         out,
			server:      server,
			traceID:     traceID,
			clientName:  clientName,
			clientAddr:  clientAddr,
			serverName:  serverName,
			serverAddr:  serverAddr,
			classMethod: classMethod,
		}
	} else {
		return &verboseTrace{
			out:         out,
			server:      server,
			traceID:     traceID,
			clientName:  clientName,
			clientAddr:  clientAddr,
			serverName:  serverName,
			serverAddr:  serverAddr,
			classMethod: classMethod,
		}
	}
}

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}
