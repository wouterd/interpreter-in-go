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
	"monkey/serializer"
	"monkey/vm"
	"os"
	"os/user"
	"path"
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

	case COMPILE:
		filename := flag.Arg(1)
		if filename == "" {
			fmt.Println("the compile command requires you to specify a file to compile.")
			fmt.Println("Like this: monkey script chimp.monkey")
			os.Exit(-1)
		}

		compileScript(filename)

	case RUN:
		filename := flag.Arg(1)
		if filename == "" {
			fmt.Println("the compile command requires you to specify a file to compile.")
			fmt.Println("Like this: monkey script chimp.monkey")
			os.Exit(-1)
		}

		runBytecode(filename)

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

func runBytecode(filename string) {
	if _, err := os.Stat(filename); err != nil {
		// filename not found, try with .mky extension
		if _, err := os.Stat(filename + ".mky"); err != nil {
			fmt.Println(fmt.Sprintf("Can't find '%s(.mky)'", filename))
			os.Exit(-1)
		}
		filename = filename + ".mky"
	}

	contents, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Can't read from", filename, ":", err.Error())
		os.Exit(-1)
	}

	l := serializer.NewLoader(contents)
	bytecode, err := l.Load()
	if err != nil {
		fmt.Println("Loaded data from ", filename)
		fmt.Println("Error loading program:", err)
		os.Exit(-1)
	}

	vm := vm.New(bytecode.Instructions, bytecode.Constants)
	err = vm.Run()
	if err != nil {
		fmt.Println("Runtime error:", err.Error())
		os.Exit(-1)
	}
}

func compileScript(filename string) {
	if _, err := os.Stat(filename); err != nil {
		// filename not found, try with .mky extension
		if _, err := os.Stat(filename + ".monkey"); err != nil {
			fmt.Println(fmt.Sprintf("Can't find '%s(.monkey)'", filename))
			os.Exit(-1)
		}
		filename = filename + ".monkey"
	}

	base := path.Base(filename)
	ext := path.Ext(filename)
	outFile := path.Join(path.Dir(filename), base[:len(base)-len(ext)]) + ".mky"
	fmt.Println("compiling", filename, "into", outFile, "...")

	bytecode := loadScript(filename)
	s := serializer.New()
	s.Write(bytecode)

	if err := os.WriteFile(outFile, s.Output, 0666); err != nil {
		fmt.Println("Error writing results: ", err.Error())
		os.Exit(-1)
	}
}

func loadScript(filename string) *compiler.Bytecode {
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
		os.Exit(-1)
	}

	return c.Bytecode()
}

func runScript(filename string) {
	if _, err := os.Stat(filename); err != nil {
		// filename not found, try with .mky extension
		if _, err := os.Stat(filename + ".monkey"); err != nil {
			fmt.Println(fmt.Sprintf("Can't find '%s(.monkey)'", filename))
			os.Exit(-1)
		}
		filename = filename + ".monkey"
	}

	bytecode := loadScript(filename)

	vm := vm.New(bytecode.Instructions, bytecode.Constants)
	err := vm.Run()

	if err != nil {
		fmt.Println("Error in execution: ", err.Error())
		os.Exit(-1)
	}
}
