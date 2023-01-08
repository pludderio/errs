package errors

import (
	"strconv"
	"strings"
)

const (
	start = iota
	seek
	parsing
	done
)

type uncaughtPanic struct{ message string }

func (p uncaughtPanic) Error() string {
	return p.message
}

// ParsePanic allows you to get an error object from the output of a go program
// that panicked. This is particularly useful with https://github.com/mitchellh/panicwrap.
func ParsePanic(text string) (*Error, error) {
	lines := strings.Split(text, "\n")

	state := start

	var message string
	var stack []StackFrame

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		switch {
		case state == start:
			if hasPanicPrefix(line) {
				message = strings.TrimPrefix(line, "panic: ")
				state = seek
				continue
			}
			return nil, errorf("bugsnag.panicParser: Invalid line (no prefix): %s", line)
		case state == seek:
			if hasGoroutinePrefix(line) && hasRunningSuffix(line) {
				state = parsing
				continue
			}
		case state == parsing:
			if line == "" {
				state = done
				break
			}
			createdBy := hasCreatedByPrefix(line)
			line = strings.TrimPrefix(line, "created by ")

			i++

			if i >= len(lines) {
				return nil, errorf("bugsnag.panicParser: Invalid line (unpaired): %s", line)
			}

			frame, err := parsePanicFrame(line, lines[i], createdBy)
			if err != nil {
				return nil, err
			}

			stack = append(stack, *frame)
			if createdBy {
				state = done
				break
			}
		case state == done || state == parsing:
			return &Error{Err: uncaughtPanic{message}, frames: stack}, nil
		}

	}
	return nil, errorf("could not parse panic: %v", text)
}

func hasCreatedByPrefix(line string) bool {
	return strings.HasPrefix(line, "created by ")
}

func hasPanicPrefix(line string) bool {
	return strings.HasPrefix(line, "panic: ")
}

func hasGoroutinePrefix(line string) bool {
	return strings.HasPrefix(line, "goroutine ")
}

func hasRunningSuffix(line string) bool {
	return strings.HasSuffix(line, "[running]:")
}

// The lines we're passing look like this:
//
//	main.(*foo).destruct(0xc208067e98)
//	        /0/go/src/github.com/bugsnag/bugsnag-go/pan/main.go:22 +0x151
func parsePanicFrame(name string, line string, createdBy bool) (*StackFrame, error) {
	idx := strings.LastIndex(name, "(")
	if idx == -1 && !createdBy {
		return nil, errorf("bugsnag.panicParser: Invalid line (no call): %s", name)
	}
	if idx != -1 {
		name = name[:idx]
	}
	pkg := ""

	if lastSlash := strings.LastIndex(name, "/"); lastSlash >= 0 {
		pkg += name[:lastSlash] + "/"
		name = name[lastSlash+1:]
	}
	if period := strings.Index(name, "."); period >= 0 {
		pkg += name[:period]
		name = name[period+1:]
	}

	name = strings.ReplaceAll(name, "·", ".")

	if !strings.HasPrefix(line, "\t") {
		return nil, errorf("bugsnag.panicParser: Invalid line (no tab): %s", line)
	}

	idx = strings.LastIndex(line, ":")
	if idx == -1 {
		return nil, errorf("bugsnag.panicParser: Invalid line (no line number): %s", line)
	}
	file := line[1:idx]

	number := line[idx+1:]
	if idx = strings.Index(number, " +"); idx > -1 {
		number = number[:idx]
	}

	lno, err := strconv.ParseInt(number, 10, 32)
	if err != nil {
		return nil, errorf("bugsnag.panicParser: Invalid line (bad line number): %s", line)
	}

	return &StackFrame{
		File:       file,
		LineNumber: int(lno),
		Package:    pkg,
		Name:       name,
	}, nil
}
