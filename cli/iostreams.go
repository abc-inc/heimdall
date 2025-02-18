// Copyright 2025 The Heimdall authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	ghTerm "github.com/cli/go-gh/v2/pkg/term"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

const DefaultWidth = 80

// ErrClosedPagerPipe is the error returned when writing to a pager that has been closed.
type ErrClosedPagerPipe struct {
	error
}

type fileWriter interface {
	io.Writer
	Fd() uintptr
}

type fileReader interface {
	io.ReadCloser
	Fd() uintptr
}

type term interface {
	IsTerminalOutput() bool
	IsColorEnabled() bool
	Is256ColorSupported() bool
	IsTrueColorSupported() bool
	Theme() string
	Size() (int, int, error)
}

type IOStreams struct {
	term term

	In     fileReader
	Out    fileWriter
	ErrOut fileWriter

	terminalTheme string

	progressIndicatorEnabled bool
	progressIndicator        *spinner.Spinner
	progressIndicatorMu      sync.Mutex

	stdinTTYOverride  bool
	stdinIsTTY        bool
	stdoutTTYOverride bool
	stdoutIsTTY       bool
	stderrTTYOverride bool
	stderrIsTTY       bool

	colorOverride bool
	colorEnabled  bool

	neverPrompt bool

	TempFileOverride *os.File
}

func (s *IOStreams) ColorEnabled() bool {
	if s.colorOverride {
		return s.colorEnabled
	}
	return s.term.IsColorEnabled()
}

func (s *IOStreams) ColorSupport256() bool {
	if s.colorOverride {
		return s.colorEnabled
	}
	return s.term.Is256ColorSupported()
}

func (s *IOStreams) HasTrueColor() bool {
	if s.colorOverride {
		return s.colorEnabled
	}
	return s.term.IsTrueColorSupported()
}

// DetectTerminalTheme is a utility to call before starting the output pager so that the terminal background
// can be reliably detected.
func (s *IOStreams) DetectTerminalTheme() {
	if !s.ColorEnabled() {
		s.terminalTheme = "none"
		return
	}

	style := os.Getenv("GLAMOUR_STYLE")
	if style != "" && style != "auto" {
		// ensure GLAMOUR_STYLE takes precedence over "light" and "dark" themes
		s.terminalTheme = "none"
		return
	}

	s.terminalTheme = s.term.Theme()
}

// TerminalTheme returns "light", "dark", or "none" depending on the background color of the terminal.
func (s *IOStreams) TerminalTheme() string {
	if s.terminalTheme == "" {
		s.DetectTerminalTheme()
	}

	return s.terminalTheme
}

func (s *IOStreams) SetColorEnabled(colorEnabled bool) {
	s.colorOverride = true
	s.colorEnabled = colorEnabled
}

func (s *IOStreams) SetStdinTTY(isTTY bool) {
	s.stdinTTYOverride = true
	s.stdinIsTTY = isTTY
}

func (s *IOStreams) IsStdinTTY() bool {
	if s.stdinTTYOverride {
		return s.stdinIsTTY
	}
	if stdin, ok := s.In.(*os.File); ok {
		return isTerminal(stdin)
	}
	return false
}

func (s *IOStreams) SetStdoutTTY(isTTY bool) {
	s.stdoutTTYOverride = true
	s.stdoutIsTTY = isTTY
}

func (s *IOStreams) IsStdoutTTY() bool {
	if s.stdoutTTYOverride {
		return s.stdoutIsTTY
	}
	// support GH_FORCE_TTY
	if s.term.IsTerminalOutput() {
		return true
	}
	stdout, ok := s.Out.(*os.File)
	return ok && isCygwinTerminal(stdout.Fd())
}

func (s *IOStreams) SetStderrTTY(isTTY bool) {
	s.stderrTTYOverride = true
	s.stderrIsTTY = isTTY
}

func (s *IOStreams) IsStderrTTY() bool {
	if s.stderrTTYOverride {
		return s.stderrIsTTY
	}
	if stderr, ok := s.ErrOut.(*os.File); ok {
		return isTerminal(stderr)
	}
	return false
}

func (s *IOStreams) CanPrompt() bool {
	if s.neverPrompt {
		return false
	}

	return s.IsStdinTTY() && s.IsStdoutTTY()
}

func (s *IOStreams) GetNeverPrompt() bool {
	return s.neverPrompt
}

func (s *IOStreams) SetNeverPrompt(v bool) {
	s.neverPrompt = v
}

func (s *IOStreams) StartProgressIndicator() {
	s.StartProgressIndicatorWithLabel("")
}

func (s *IOStreams) StartProgressIndicatorWithLabel(label string) {
	if !s.progressIndicatorEnabled {
		return
	}

	s.progressIndicatorMu.Lock()
	defer s.progressIndicatorMu.Unlock()

	if s.progressIndicator != nil {
		if label == "" {
			s.progressIndicator.Prefix = ""
		} else {
			s.progressIndicator.Prefix = label + " "
		}
		return
	}

	// https://github.com/briandowns/spinner#available-character-sets
	dotStyle := spinner.CharSets[11]
	sp := spinner.New(dotStyle, 120*time.Millisecond, spinner.WithWriter(s.ErrOut), spinner.WithColor("fgCyan"))
	if label != "" {
		sp.Prefix = label + " "
	}

	sp.Start()
	s.progressIndicator = sp
}

func (s *IOStreams) StopProgressIndicator() {
	s.progressIndicatorMu.Lock()
	defer s.progressIndicatorMu.Unlock()
	if s.progressIndicator == nil {
		return
	}
	s.progressIndicator.Stop()
	s.progressIndicator = nil
}

func (s *IOStreams) RunWithProgress(label string, run func() error) error {
	s.StartProgressIndicatorWithLabel(label)
	defer s.StopProgressIndicator()

	return run()
}

func (s *IOStreams) RefreshScreen() {
	if s.IsStdoutTTY() {
		// Move cursor to 0,0
		fmt.Fprint(s.Out, "\x1b[0;0H")
		// Clear from cursor to bottom of screen
		fmt.Fprint(s.Out, "\x1b[J")
	}
}

// TerminalWidth returns the width of the terminal that controls the process
func (s *IOStreams) TerminalWidth() int {
	w, _, err := s.term.Size()
	if err == nil && w > 0 {
		return w
	}
	return DefaultWidth
}

func (s *IOStreams) ReadUserFile(fn string) ([]byte, error) {
	var r io.ReadCloser
	if fn == "-" {
		r = s.In
	} else {
		var err error
		r, err = os.Open(fn)
		if err != nil {
			return nil, err
		}
	}
	defer r.Close()
	return io.ReadAll(r)
}

func (s *IOStreams) TempFile(dir, pattern string) (*os.File, error) {
	if s.TempFileOverride != nil {
		return s.TempFileOverride, nil
	}
	return os.CreateTemp(dir, pattern)
}

func System() *IOStreams {
	terminal := ghTerm.FromEnv()

	var stdout fileWriter = os.Stdout
	// On Windows with no virtual terminal processing support, translate ANSI escape
	// sequences to console syscalls.
	if colorableStdout := colorable.NewColorable(os.Stdout); colorableStdout != os.Stdout {
		// Ensure that the file descriptor of the original stdout is preserved.
		stdout = &fdWriter{
			fd:     os.Stdout.Fd(),
			Writer: colorableStdout,
		}
	}

	var stderr fileWriter = os.Stderr
	// On Windows with no virtual terminal processing support, translate ANSI escape
	// sequences to console syscalls.
	if colorableStderr := colorable.NewColorable(os.Stderr); colorableStderr != os.Stderr {
		// Ensure that the file descriptor of the original stderr is preserved.
		stderr = &fdWriter{
			fd:     os.Stderr.Fd(),
			Writer: colorableStderr,
		}
	}

	io := &IOStreams{
		In:     os.Stdin,
		Out:    stdout,
		ErrOut: stderr,
		term:   &terminal,
	}

	stdoutIsTTY := io.IsStdoutTTY()
	stderrIsTTY := io.IsStderrTTY()

	if stdoutIsTTY && stderrIsTTY {
		io.progressIndicatorEnabled = true
	}

	return io
}

type fakeTerm struct{}

func (t fakeTerm) IsTerminalOutput() bool {
	return false
}

func (t fakeTerm) IsColorEnabled() bool {
	return false
}

func (t fakeTerm) Is256ColorSupported() bool {
	return false
}

func (t fakeTerm) IsTrueColorSupported() bool {
	return false
}

func (t fakeTerm) Theme() string {
	return ""
}

func (t fakeTerm) Size() (int, int, error) {
	return 80, -1, nil
}

func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	writer = nil

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	io := &IOStreams{
		In: &fdReader{
			fd:         0,
			ReadCloser: io.NopCloser(in),
		},
		Out:    &fdWriter{fd: 1, Writer: out},
		ErrOut: &fdWriter{fd: 2, Writer: errOut},
		term:   &fakeTerm{},
	}
	io.SetStdinTTY(false)
	io.SetStdoutTTY(false)
	io.SetStderrTTY(false)
	return io, in, out, errOut
}

func isTerminal(f *os.File) bool {
	return ghTerm.IsTerminal(f) || isCygwinTerminal(f.Fd())
}

func isCygwinTerminal(fd uintptr) bool {
	return isatty.IsCygwinTerminal(fd)
}

// fdWriter represents a wrapped stdout Writer that preserves the original file descriptor
type fdWriter struct {
	io.Writer
	fd uintptr
}

func (w *fdWriter) Fd() uintptr {
	return w.fd
}

// fdWriteCloser represents a wrapped stdout Writer that preserves the original file descriptor
type fdWriteCloser struct {
	io.WriteCloser
	fd uintptr
}

func (w *fdWriteCloser) Fd() uintptr {
	return w.fd
}

// fdReader represents a wrapped stdin ReadCloser that preserves the original file descriptor
type fdReader struct {
	io.ReadCloser
	fd uintptr
}

func (r *fdReader) Fd() uintptr {
	return r.fd
}
