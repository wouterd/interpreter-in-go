package main

import (
	"flag"
	"fmt"
	"monkey/compiler"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/repl"
	"monkey/vm"
	"os"
	"os/user"
)

type Mode int

const (
	REPL = iota
	COMPILE
	RUN
	SCRIPT
)

func main() {
	flag.Parse()

	mode := REPL

	if len(flag.Args()) > 0 {
		switch flag.Arg(0) {
		case "c", "compile":
			mode = COMPILE
		case "r", "run":
			mode = RUN
		case "s", "script":
			mode = SCRIPT
		case "repl", "eval", "console":
			mode = REPL
		default:
			fmt.Println(fmt.Sprintf("Unknown command: %s", flag.Arg(0)))
			os.Exit(-1)
		}
	}

	switch mode {
	case REPL:
		user, err := user.Current()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)
		fmt.Printf("Feel free to type in commands\n")
		repl.Start(os.Stdin, os.Stdout)
		os.Exit(0)

	case SCRIPT:
		filename := flag.Arg(1)
		if filename == "" {
			fmt.Println("the script command requires you to specify a script file to execute.")
			fmt.Println("Like this: monkey script chimp.monkey")
			os.Exit(-1)
		}

		runScript(filename)

	default:
		fmt.Println("Not implemented yet...")
		os.Exit(-1)
	}

	os.Exit(0)
}

func runScript(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(fmt.Sprintf("Can't open file %s: %s", filename, err.Error()))
		os.Exit(-1)
	}

	stat, err := file.Stat()
	if err != nil {
		fmt.Println("Can't stat the script")
		os.Exit(-1)
	}

	buf := make([]byte, stat.Size())
	_, err = file.Read(buf)
	if err != nil {
		fmt.Println("Can't read file contents: ", err.Error())
		os.Exit(-1)
	}

	l := lexer.New(string(buf))
	p := parser.New(l)
	program := p.ParseProgram()

	macroEnv := object.NewEnvironment()
	evaluator.DefineMacros(program, macroEnv)
	expanded := evaluator.ExpandMacros(program, macroEnv)

	if len(p.Errors()) > 0 {
		fmt.Println("Error(s) parsing the script:")
		for _, e := range p.Errors() {
			fmt.Println(e)
		}
		os.Exit(-1)
	}

	c := compiler.New()
	err = c.Compile(expanded)

	if err != nil {
		fmt.Println("Error while compiling script: ", err.Error())
	}

	vm := vm.New(c.Bytecode().Instructions, c.Bytecode().Constants)
	err = vm.Run()

	if err != nil {
		fmt.Println("Error in execution: ", err.Error())
		os.Exit(-1)
	}
}
