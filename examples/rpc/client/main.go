package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/ironzhang/tlog"
	"github.com/ironzhang/zerone/examples/rpc/arith"
	"github.com/ironzhang/zerone/rpc"
)

type Options struct {
	net  string
	addr string
}

func (o *Options) Parse() {
	flag.StringVar(&o.net, "net", "tcp", "network")
	flag.StringVar(&o.addr, "addr", ":10000", "address")
	flag.Parse()
}

type Command struct {
	name string
	args []string
}

func ParseCommand(line string) (*Command, error) {
	cmds := strings.Split(strings.TrimSpace(line), " ")
	if len(cmds) <= 0 {
		return nil, fmt.Errorf("%q is a invalid command", line)
	}
	return &Command{
		name: cmds[0],
		args: cmds[1:],
	}, nil
}

func (p *Command) Execute(e *Executer) error {
	switch p.name {
	case "add":
		return e.ArithAdd(p.args)
	case "sub":
		return e.ArithSub(p.args)
	case "mul":
		return e.ArithMul(p.args)
	case "div":
		return e.ArithDiv(p.args)
	case "verbose":
		return e.Verbose(p.args)
	case "quit":
		fmt.Printf("bye\n")
		os.Exit(0)
	default:
		return fmt.Errorf("%q is a unknowm command", p.name)
	}
	return nil
}

type Executer struct {
	verbose int
	client  *rpc.Client
}

func (p *Executer) ArithAdd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("invalid params")
	}
	a, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	b, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	var reply int
	if err := p.client.Call(rpc.WithVerbose(context.Background(), p.verbose), "Arith.Add", arith.Args{a, b}, &reply, 0); err != nil {
		return err
	}
	fmt.Printf("%d\n", reply)
	return nil
}

func (p *Executer) ArithSub(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("invalid params")
	}
	a, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	b, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	var reply int
	if err := p.client.Call(rpc.WithVerbose(context.Background(), p.verbose), "Arith.Sub", arith.Args{a, b}, &reply, 0); err != nil {
		return err
	}
	fmt.Printf("%d\n", reply)
	return nil
}

func (p *Executer) ArithMul(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("invalid params")
	}
	a, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	b, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	var reply int
	if err := p.client.Call(rpc.WithVerbose(context.Background(), p.verbose), "Arith.Multiply", arith.Args{a, b}, &reply, 0); err != nil {
		return err
	}
	fmt.Printf("%d\n", reply)
	return nil
}

func (p *Executer) ArithDiv(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("invalid params")
	}
	a, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	b, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	var quo arith.Quotient
	if err := p.client.Call(rpc.WithVerbose(context.Background(), p.verbose), "Arith.Divide", arith.Args{a, b}, &quo, 0); err != nil {
		return err
	}
	fmt.Printf("quo: %d, rem: %d\n", quo.Quo, quo.Rem)
	return nil
}

func (p *Executer) Verbose(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("invalid params")
	}
	verbose, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	switch args[0] {
	case "trace":
		p.verbose = verbose
	case "client":
		p.client.SetTraceVerbose(verbose)
	case "server":
		if err := p.client.Call(context.Background(), "Trace.SetVerbose", verbose, nil, 0); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid params")
	}
	return nil
}

func main() {
	var opts Options
	opts.Parse()

	c, err := rpc.Dial("ArithClient", opts.net, opts.addr)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer c.Close()

	e := &Executer{client: c}
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("$")
		line, err := r.ReadString('\n')
		if err != nil {
			log.Fatalf("read string: %v", err)
		}
		cmd, err := ParseCommand(line)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		if err = cmd.Execute(e); err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
	}
}
