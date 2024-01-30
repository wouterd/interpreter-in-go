package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/compiler"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
	"os"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	//env := object.NewEnvironment()
	macroEnv := object.NewEnvironment()

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		if line[0] == ':' {
			i := 1
			for i < len(line) && line[i] != ' ' {
				i++
			}

			cmd := line[1:i]

			switch cmd {
			case "load":
				args := line[i+1:]
				if len(args) > 0 {
					cnts, err := os.ReadFile(args)
					if err != nil {
						io.WriteString(out, fmt.Sprintf("ERROR reading %s: %s\n", args, err.Error()))
						continue
					}
					line = string(cnts)
				} else {
					io.WriteString(out, "USAGE: load [filename]\n")
					continue
				}
			case "macros":
				for name, obj := range macroEnv.All() {
					macro, ok := obj.(*object.Macro)
					if ok {
						fmt.Printf("%s: %s\n", name, macro.Inspect())
					}
				}
				continue
			}
		}

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluator.DefineMacros(program, macroEnv)
		expanded := evaluator.ExpandMacros(program, macroEnv)

		comp := compiler.New()
		err := comp.Compile(expanded)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		machine := vm.New(comp.Bytecode())
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
		}

		stackTop := machine.LastPoppedStackElem()

		if stackTop != nil {
			io.WriteString(out, stackTop.Inspect())
			io.WriteString(out, "\n")
		}

		//		evaluated := evaluator.Eval(expanded, env)
		//
		//		if evaluated != nil {
		//			io.WriteString(out, evaluated.Inspect())
		//			io.WriteString(out, "\n")
		//		}
	}

}

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
