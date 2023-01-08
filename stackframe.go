package errors

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// A StackFrame contains all necessary information about to generate a line
// in a callstack.
type StackFrame struct {
	// The path to the file containing this ProgramCounter
	File string
	// The LineNumber in that file
	LineNumber int
	// The Name of the function that contains this ProgramCounter
	Name string
	// The Package that contains this function
	Package string
	// The underlying ProgramCounter
	ProgramCounter uintptr
}

// NewStackFrame populates a stacks frame object from the program counter.
func NewStackFrame(programCounter uintptr) StackFrame {

	frame := StackFrame{ProgramCounter: programCounter}
	if frame.Func() == nil {
		return frame
	}
	frame.Package, frame.Name = packageAndName(frame.Func())

	// pc -1 because the program counters we use are usually return addresses,
	// and we want to show the line that corresponds to the function call
	frame.File, frame.LineNumber = frame.Func().FileLine(programCounter - 1)
	return frame

}

// Func returns the function that contained this frame.
func (frame *StackFrame) Func() *runtime.Func {
	if frame.ProgramCounter == 0 {
		return nil
	}
	return runtime.FuncForPC(frame.ProgramCounter)
}

// String returns the stackframe formatted in the same way as go does
// in runtime/debug.Stack().
func (frame *StackFrame) String() string {
	str := fmt.Sprintf("%s:%d (0x%x)\n", frame.File, frame.LineNumber, frame.ProgramCounter)

	source, err := frame.sourceLine()
	if err != nil {
		return str
	}

	return str + fmt.Sprintf("\t%s: %s\n", frame.Name, source)
}

// SourceLine gets the line of code (from File and Line) of the original source if possible.
func (frame *StackFrame) SourceLine() (string, error) {
	source, err := frame.sourceLine()
	if err != nil {
		return source, new(err)
	}
	return source, err
}

func (frame *StackFrame) sourceLine() (string, error) {
	if frame.LineNumber <= 0 {
		return "???", nil
	}

	file, err := os.Open(frame.File)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine == frame.LineNumber {
			return string(bytes.Trim(scanner.Bytes(), " \t")), nil
		}
		currentLine++
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "???", nil
}

func packageAndName(fn *runtime.Func) (string, string) {
	name := fn.Name()
	pkg := ""

	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//  runtime/debug.*T·ptrmethod
	// and want
	//  *T.ptrmethod
	// Since the package path might contains dots (e.g. code.google.com/...),
	// we first remove the path prefix if there is one.
	if lastslash := strings.LastIndex(name, "/"); lastslash >= 0 {
		pkg += name[:lastslash] + "/"
		name = name[lastslash+1:]
	}
	if period := strings.Index(name, "."); period >= 0 {
		pkg += name[:period]
		name = name[period+1:]
	}

	name = strings.ReplaceAll(name, "·", ".")
	return pkg, name
}
