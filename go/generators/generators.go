// Package generators is a Go port of smc.generators. A CodeGenerator renders
// an optimized state machine with a language's implementer and writes the
// resulting files.
package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/jeb2239/CC_SMC/go/implementers"
	"github.com/jeb2239/CC_SMC/go/nsc"
	"github.com/jeb2239/CC_SMC/go/optimizer"
)

// CodeGenerator generates the source files for one state machine.
type CodeGenerator interface {
	Generate() error
}

type constructor func(sm *optimizer.StateMachine, outputDirectory string, flags map[string]string) CodeGenerator

// Java finds generators by class name, smc.generators.<language>CodeGenerator;
// this table plays that part.
var generators = map[string]constructor{
	"C":    NewCCodeGenerator,
	"Cpp":  NewCppCodeGenerator,
	"Dart": NewDartCodeGenerator,
	"Go":   NewGoCodeGenerator,
	"Java": NewJavaCodeGenerator,
}

// Languages returns the names New accepts, sorted.
func Languages() []string {
	languages := make([]string, 0, len(generators))
	for language := range generators {
		languages = append(languages, language)
	}
	slices.Sort(languages)
	return languages
}

// New returns the generator for language. Files are written to
// outputDirectory, or to the current directory if it is empty.
func New(language string, sm *optimizer.StateMachine, outputDirectory string, flags map[string]string) (CodeGenerator, error) {
	newGenerator, ok := generators[language]
	if !ok {
		return nil, fmt.Errorf("no code generator for language %q; choose one of %s",
			language, strings.Join(Languages(), ", "))
	}
	return newGenerator(sm, outputDirectory, flags), nil
}

type codeGenerator struct {
	sm              *optimizer.StateMachine
	outputDirectory string
	flags           map[string]string
}

func (g *codeGenerator) outputPath(outputFileName string) string {
	return filepath.Join(g.outputDirectory, outputFileName)
}

func (g *codeGenerator) render(implementer nsc.Visitor) {
	nsc.Generate(g.sm).Accept(implementer)
}

func (g *codeGenerator) writeFile(outputFileName, content string) error {
	return os.WriteFile(g.outputPath(outputFileName), []byte(content), 0o644)
}

type JavaCodeGenerator struct {
	codeGenerator
	implementer *implementers.JavaNestedSwitchCaseImplementer
}

func NewJavaCodeGenerator(sm *optimizer.StateMachine, outputDirectory string, flags map[string]string) CodeGenerator {
	return &JavaCodeGenerator{
		codeGenerator{sm, outputDirectory, flags},
		implementers.NewJavaNestedSwitchCaseImplementer(flags),
	}
}

func (g *JavaCodeGenerator) Generate() error {
	g.render(g.implementer)
	return g.writeFile(g.sm.Header.Fsm+".java", g.implementer.Output())
}

type CCodeGenerator struct {
	codeGenerator
	implementer *implementers.CNestedSwitchCaseImplementer
}

func NewCCodeGenerator(sm *optimizer.StateMachine, outputDirectory string, flags map[string]string) CodeGenerator {
	return &CCodeGenerator{
		codeGenerator{sm, outputDirectory, flags},
		implementers.NewCNestedSwitchCaseImplementer(flags),
	}
}

func (g *CCodeGenerator) Generate() error {
	g.render(g.implementer)
	if errors := g.implementer.Errors(); len(errors) > 0 {
		for _, err := range errors {
			fmt.Println("Implementation error: " + err.String())
		}
		return nil
	}
	fileName := toLowerCamelCase(g.sm.Header.Fsm)
	if err := g.writeFile(fileName+".h", g.implementer.FsmHeader()); err != nil {
		return err
	}
	return g.writeFile(fileName+".c", g.implementer.FsmImplementation())
}

func toLowerCamelCase(s string) string {
	if len(s) < 2 {
		return strings.ToLower(s)
	}
	return strings.ToLower(s[:1]) + s[1:]
}

type CppCodeGenerator struct {
	codeGenerator
	implementer *implementers.CppNestedSwitchCaseImplementer
}

func NewCppCodeGenerator(sm *optimizer.StateMachine, outputDirectory string, flags map[string]string) CodeGenerator {
	return &CppCodeGenerator{
		codeGenerator{sm, outputDirectory, flags},
		implementers.NewCppNestedSwitchCaseImplementer(flags),
	}
}

func (g *CppCodeGenerator) Generate() error {
	g.render(g.implementer)
	return g.writeFile(g.sm.Header.Fsm+".h", g.implementer.Output())
}

type DartCodeGenerator struct {
	codeGenerator
	implementer *implementers.DartNestedSwitchCaseImplementer
}

func NewDartCodeGenerator(sm *optimizer.StateMachine, outputDirectory string, flags map[string]string) CodeGenerator {
	return &DartCodeGenerator{
		codeGenerator{sm, outputDirectory, flags},
		implementers.NewDartNestedSwitchCaseImplementer(flags),
	}
}

func (g *DartCodeGenerator) Generate() error {
	g.render(g.implementer)
	return g.writeFile(g.sm.Header.Fsm+".dart", g.implementer.Output())
}

type GoCodeGenerator struct {
	codeGenerator
	implementer *implementers.GoNestedSwitchCaseImplementer
}

func NewGoCodeGenerator(sm *optimizer.StateMachine, outputDirectory string, flags map[string]string) CodeGenerator {
	return &GoCodeGenerator{
		codeGenerator{sm, outputDirectory, flags},
		implementers.NewGoNestedSwitchCaseImplementer(flags),
	}
}

func (g *GoCodeGenerator) Generate() error {
	g.render(g.implementer)
	return g.writeFile(camelToSnake(g.sm.Header.Fsm+".go"), g.implementer.Output())
}

var camelHump = regexp.MustCompile(`([a-z])([A-Z]+)`)

func camelToSnake(s string) string {
	return strings.ToLower(camelHump.ReplaceAllString(s, "${1}_${2}"))
}
