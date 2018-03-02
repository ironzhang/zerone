package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ironzhang/zerone/examples/rpc/arith"
	"github.com/ironzhang/zerone/rpc"
	"github.com/ironzhang/zerone/zlog"
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

func (p *Command) Execute(c *rpc.Client) error {
	switch p.name {
	case "Arith.Add":
		return ArithAdd(c, p.args)
	case "Arith.Sub":
	case "Arith.Multiply":
	case "Arith.Divide":
	case "quit":
		fmt.Printf("bye\n")
		os.Exit(0)
	default:
		return fmt.Errorf("%q is a unknowm command", p.name)
	}
	return nil
}

func ArithAdd(c *rpc.Client, args []string) error {
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
	if err := c.Call(context.Background(), "Arith.Add", arith.Args{a, b}, &reply); err != nil {
		return err
	}
	fmt.Printf("%d\n", reply)
	return nil
}

func main() {
	var opts Options
	opts.Parse()

	c, err := rpc.Dial("ArithClient", opts.net, opts.addr)
	if err != nil {
		zlog.Fatalf("dial: %v", err)
	}
	defer c.Close()

	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("$")
		line, err := r.ReadString('\n')
		if err != nil {
			zlog.Fatalf("read string: %v", err)
		}
		cmd, err := ParseCommand(line)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		if err = cmd.Execute(c); err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
	}
}
