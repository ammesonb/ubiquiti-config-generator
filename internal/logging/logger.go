package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"golang.org/x/term"
)

// Logger writes timestamped terminal messages without changing process exit behavior.
type Logger struct {
	output io.Writer
	color  bool
	now    func() time.Time
	mu     sync.Mutex
}

// SupportsColor respects terminal capabilities and the user's NO_COLOR preference.
func SupportsColor(output io.Writer) bool {
	file, ok := output.(*os.File)
	return ok && term.IsTerminal(int(file.Fd())) && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
}

func New(output io.Writer) *Logger {
	enabled := SupportsColor(output)
	if file, ok := output.(*os.File); ok && enabled {
		output = colorable.NewColorable(file)
	}
	return &Logger{output: output, color: enabled, now: time.Now}
}

func (l *Logger) Info(message string) error     { return l.write("INFO", color.FgCyan, message) }
func (l *Logger) Warn(message string) error     { return l.write("WARN", color.FgYellow, message) }
func (l *Logger) Error(message string) error    { return l.write("ERR", color.FgRed, message) }
func (l *Logger) Critical(message string) error { return l.write("CRIT", color.FgHiRed, message) }

func (l *Logger) write(level string, attribute color.Attribute, message string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	timestamp := l.now().UTC().Format(time.RFC3339)
	var record strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(message, "\n"), "\n") {
		text := fmt.Sprintf("%s [%4s] %s", timestamp, level, line)
		if l.color {
			style := color.New(attribute)
			style.EnableColor()
			text = style.Sprint(text)
		}
		record.WriteString(text)
		record.WriteByte('\n')
	}
	_, err := io.WriteString(l.output, record.String())
	return err
}
