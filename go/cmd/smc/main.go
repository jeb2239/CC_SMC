// Command smc is the State Machine Compiler, a Go port of smc.SMC.
//
//	smc [-l language] [-o directory] [-f flags] file
//
// language is one of C, Cpp, Dart, Go or Java (the default). Output goes to
// directory, or to the current directory. flags is a list of key:value pairs
// separated by commas; Java understands package:<name>.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jeb2239/CC_SMC/go/args"
	"github.com/jeb2239/CC_SMC/go/generators"
	"github.com/jeb2239/CC_SMC/go/lexer"
	"github.com/jeb2239/CC_SMC/go/optimizer"
	"github.com/jeb2239/CC_SMC/go/parser"
	"github.com/jeb2239/CC_SMC/go/semantic"
)

const argSchema = "o*,l*,f&"

func main() {
	argParser, err := args.New(argSchema, os.Args[1:])
	if err != nil {
		usage(err.Error())
	}
	if err := newSmcCompiler(os.Args[1:], argParser).run(); err != nil {
		if !errors.Is(err, errCompilation) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

// errCompilation reports that the source had errors, which run has already
// printed. SMC.java exits with status 0 in that case.
var errCompilation = errors.New("compilation failed")

func usage(message string) {
	fmt.Println("usage: " + argSchema + " file")
	fmt.Println(message)
	os.Exit(2)
}

type smcCompiler struct {
	args            []string
	argParser       *args.Args
	outputDirectory string
	language        string
	flags           map[string]string
}

func newSmcCompiler(arguments []string, argParser *args.Args) *smcCompiler {
	return &smcCompiler{
		args:      arguments,
		argParser: argParser,
		language:  "Java",
		flags:     map[string]string{},
	}
}

func (c *smcCompiler) run() error {
	c.extractCommandLineArguments()

	source, err := c.sourceCode()
	if err != nil {
		return err
	}
	fsm := compile(source)
	if reportSyntaxErrors(fsm) != 0 {
		return errCompilation
	}

	ast := semantic.NewAnalyzer().Analyze(fsm)
	if reportSemanticErrors(ast) != 0 {
		return errCompilation
	}
	return c.generateCode(optimizer.Optimize(ast))
}

func (c *smcCompiler) extractCommandLineArguments() {
	if c.argParser.Has('o') {
		c.outputDirectory = c.argParser.GetString('o')
	}
	if c.argParser.Has('l') {
		c.language = c.argParser.GetString('l')
	}
	if c.argParser.Has('f') {
		c.flags = c.argParser.GetMap('f')
	}
}

func compile(smContent string) *parser.FsmSyntax {
	syntaxBuilder := parser.NewSyntaxBuilder()
	p := parser.New(syntaxBuilder)
	lexer.New(p).Lex(smContent)
	p.HandleEvent(parser.EventEOF, -1, -1)
	return syntaxBuilder.Fsm()
}

func (c *smcCompiler) sourceCode() (string, error) {
	next := c.argParser.NextArgument()
	if next >= len(c.args) {
		usage("No source file given.")
	}
	source, err := os.ReadFile(c.args[next])
	return string(source), err
}

func reportSyntaxErrors(fsm *parser.FsmSyntax) int {
	syntaxErrorCount := len(fsm.Errors)
	fmt.Printf("Compiled with %d syntax error%s.\n", syntaxErrorCount, plural(syntaxErrorCount))
	for _, err := range fsm.Errors {
		fmt.Println(err.Error())
	}
	return syntaxErrorCount
}

// reportSemanticErrors has no Java counterpart: SMC.java goes on to optimize
// a state machine with semantic errors, and fails with a NullPointerException.
func reportSemanticErrors(ast *semantic.StateMachine) int {
	semanticErrorCount := len(ast.Errors)
	if semanticErrorCount != 0 {
		fmt.Printf("Compiled with %d semantic error%s.\n", semanticErrorCount, plural(semanticErrorCount))
		for _, err := range ast.Errors {
			fmt.Println(err.String())
		}
	}
	return semanticErrorCount
}

func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func (c *smcCompiler) generateCode(sm *optimizer.StateMachine) error {
	generator, err := generators.New(c.language, sm, c.outputDirectory, c.flags)
	if err != nil {
		return err
	}
	return generator.Generate()
}
