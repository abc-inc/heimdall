// Copyright 2023 The Heimdall authors
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

//go:build !no_interactive

package interactive

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/alecthomas/chroma/lexers/b"
	"github.com/alecthomas/chroma/lexers/y"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/alessio/shellescape"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/muesli/termenv"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"mvdan.cc/sh/v3/shell"
)

func NewInteractiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "interactive [<args>]",
		Short:   "Interactive CLI builder",
		GroupID: console.HeimdallGroup,
		Args:    cobra.MaximumNArgs(100),
		Run: func(cmd *cobra.Command, args []string) {
			draw(cmd.Root(), args...)
		},
	}

	return cmd
}

const argsLabel = "Arguments"

var outputs = []string{"csv", "json", "table", "text", "tsv", "yaml"}
var subCmd *cobra.Command
var form *tview.Form
var docs *tview.TextView
var cmdLine *tview.TextView
var cmdLineWriter io.Writer

var fieldBG, fieldFG, sectFG = tcell.ColorSlateGray, tcell.ColorBlack, "[yellow::b]"

func draw(cmd *cobra.Command, args ...string) {
	if !termenv.HasDarkBackground() {
		tview.Styles.PrimitiveBackgroundColor = tcell.ColorWhite
		tview.Styles.ContrastBackgroundColor = tcell.ColorSilver
		tview.Styles.MoreContrastBackgroundColor = tcell.ColorDarkGray
		tview.Styles.BorderColor = tcell.ColorGray
		tview.Styles.TitleColor = tcell.ColorGray
		tview.Styles.GraphicsColor = tcell.ColorBlack
		tview.Styles.PrimaryTextColor = tcell.ColorBlack
		tview.Styles.SecondaryTextColor = tcell.ColorDarkSlateGray
		tview.Styles.TertiaryTextColor = tcell.ColorGreen
		tview.Styles.InverseTextColor = tcell.ColorBlue
		tview.Styles.ContrastSecondaryTextColor = tcell.ColorYellow

		fieldBG, fieldFG, sectFG = tcell.ColorSilver, tcell.ColorBlack, "[red::b]"
	}

	app := tview.NewApplication().EnableMouse(true)
	pages := tview.NewPages()

	w, _, _ := term.GetSize(int(os.Stdout.Fd()))
	cmdLine = tview.NewTextView().SetLabel(("Command Line")+" ").
		SetSize(tview.DefaultFormFieldHeight, w).
		SetDynamicColors(true).
		SetScrollable(true).
		SetText("")

	cmdLineWriter = tview.ANSIWriter(cmdLine)

	frame := tview.NewFrame(pages).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText("", false, tview.AlignLeft, tcell.ColorLightSlateGrey).
		AddText("Ctrl+C: Exit • Ctrl+Q: Copy to clipboard • Ctrl+R: Run • Ctrl+W: Close preview • Ctrl+X: Exit and execute command",
			false, tview.AlignLeft, tcell.ColorLightSlateGrey)

	preview := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)

	preview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			text := preview.GetText(true)
			internal.MustNoErr(clipboard.WriteAll(text))
		} else if event.Key() == tcell.KeyCtrlW {
			pages.RemovePage("preview")
			app.Sync()
			preview.SetText("")
		}
		return event
	})

	docs = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	docs.SetBorderPadding(1, 1, 1, 1)
	form = tview.NewForm().SetFieldBackgroundColor(fieldBG).SetFieldTextColor(fieldFG)
	createCmdDropdown(cmd)

	grid := tview.NewGrid().
		SetRows(-10, -10, -1).
		SetColumns(60, -1).
		SetBorders(true)

	grid.AddItem(form, 0, 0, 1, 2, 10, 10, true).
		AddItem(cmdLine, 2, 0, 1, 2, 0, 0, false)

	grid.AddItem(form, 0, 0, 2, 1, 0, 80, true).
		AddItem(docs, 0, 1, 2, 1, 0, 80, false)

	pages.AddPage("main", grid, true, true)

	if len(args) > 0 {
		if filepath.Base(args[0]) != filepath.Base(os.Args[0]) {
			args = slices.Insert(args, 0, filepath.Base(os.Args[0]))
		}
		initForm(cmd, args...)
	} else if cb, err := clipboard.ReadAll(); err == nil && cb != "" {
		initForm(cmd, strings.Fields(cb)...)
	}

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		text := cmdLine.GetText(true)
		if event.Key() == tcell.KeyCtrlQ {
			internal.MustNoErr(clipboard.WriteAll(text))
			return nil
		} else if event.Key() == tcell.KeyCtrlR && text != "" {
			_ = subCmd.PersistentFlags().Set("output", "json")
			cmd.SetArgs(internal.Must(shell.Fields(text, nil))[1:])
			pages.AddAndSwitchToPage("preview", preview, true)

			out, format := "", ""
			if outputFI := form.GetFormItemByLabel("Output"); outputFI != nil {
				_, format = outputFI.(*tview.DropDown).GetCurrentOption()
			}
			if cmd.Name() == "echo" || strings.HasSuffix(format, "c") {
				out = tview.TranslateANSI(execANSI(cmd))
			} else {
				out = execSimple(cmd)
			}
			preview.SetText(out)

			// Slice flags keep their values, so they need to be reset.
			subCmd.Flags().Visit(func(flag *pflag.Flag) {
				if v, ok := flag.Value.(pflag.SliceValue); ok {
					internal.MustNoErr(v.Replace([]string{}))
				}
			})
		} else if event.Key() == tcell.KeyCtrlX && text != "" {
			app.Stop()
			log.Info().Str("cmd", text).Msg("Executing")
			cmd.SetArgs(internal.Must(shell.Fields(text, nil))[1:])
			internal.MustNoErr(cmd.Execute())
		}
		return event
	})

	if err := app.SetRoot(frame, true).Run(); err != nil {
		panic(err)
	}
}

func initForm(cmd *cobra.Command, args ...string) {
	log.Info().Str("cmd", cmd.Name()).Msg("Initializing form")
	if len(args) > 0 && filepath.Base(args[0]) == filepath.Base(os.Args[0]) {
		if subCmd, subArgs, err := cmd.Traverse(args[1:]); err == nil {
			preSelect(subCmd, subArgs...)
		}
	}
}

func preSelect(cmd *cobra.Command, args ...string) {
	if cmd == nil || cmd == cmd.Root() {
		return
	}
	preSelect(cmd.Parent(), args...)
	dd := form.GetFormItem(form.GetFormItemCount() - 1).(*tview.DropDown)
	index := slices.Index(listCmdNames(cmd.Parent()), cmd.Name())
	dd.SetCurrentOption(index)

	argsFI := form.GetFormItemByLabel(argsLabel)
	if argsFI != nil {
		var remArgs []string
		fs := internal.Must(shell.Fields(strings.Join(args, " "), nil))
		for _, arg := range fs {
			if !strings.HasPrefix(arg, "-") {
				remArgs = append(remArgs, arg)
				continue
			}
			f, v, _ := strings.Cut(arg, "=")
			l := cases.Title(language.English).String(strings.TrimLeft(f, "-"))
			if fi := form.GetFormItemByLabel(l); fi == nil {
				remArgs = append(remArgs, arg)
			} else if c, ok := fi.(*tview.Checkbox); ok {
				c.SetChecked(true)
			} else if d, ok := fi.(*tview.DropDown); ok && l == "Output" {
				d.SetCurrentOption(slices.Index(outputs, v))
			} else {
				if !strings.HasPrefix(v, `"`) && !strings.HasPrefix(v, `'`) {
					v = shellescape.Quote(v)
				}
				fi.(*tview.InputField).SetText(v)
			}
		}
		argsFI.(*tview.InputField).SetText(strings.Join(remArgs, " "))
	}
}

func createCmdDropdown(cmd *cobra.Command) *Section {
	s := &Section{}
	s.FormItem = tview.NewDropDown().
		SetLabel((strings.Repeat("Sub-", form.GetFormItemCount())) + "Command")

	sel := func(option string, optionIndex int) {
		if _, currOpt := s.FormItem.(*tview.DropDown).GetCurrentOption(); currOpt == s.PrevVal {
			return
		}

		var err error
		subCmd, _, err = cmd.Find([]string{option})
		internal.MustNoErr(err)
		s.DisposeChildren()

		if len(listCmdNames(subCmd)) == 0 {
			s.Child = createFlags(subCmd)
		} else if option != "" {
			s.Child = createCmdDropdown(subCmd)
		}

		usg := subCmd.UsageString()
		if len(usg) > 0 {
			buf := &strings.Builder{}
			// The dynamic colors option replaces all [], so we need to replace them with <>.
			usg = strings.NewReplacer(`[`, `<`, `]`, `>`).Replace(usg)
			internal.MustNoErr(quick.Highlight(tview.ANSIWriter(buf), usg,
				y.YAML.Config().Name, "terminal", styles.Get("GitHub").Name))

			usg = strings.ReplaceAll(buf.String(), "[silver::b]", sectFG)
			docs.SetText(usg)
		}
		s.PrevVal = option
	}

	ns := listCmdNames(cmd)
	s.Reset = resetForm(form.GetFormItemCount())

	s.FormItem.(*tview.DropDown).
		SetOptions(ns, sel).
		SetCurrentOption(slices.Index(ns, cmd.Name()))
	form.AddFormItem(s.FormItem)
	return s
}

func resetForm(cnt int) func() {
	return func() {
		for i := form.GetFormItemCount() - 1; i >= cnt; i-- {
			form.RemoveFormItem(i)
		}
	}
}

func listCmdNames(cmd *cobra.Command) (ns []string) {
	ignore := []string{"completion", "help", "http", "interactive", "run"}
	for _, c := range cmd.Commands() {
		if c.Hidden || slices.Contains(ignore, c.Name()) || strings.HasPrefix(c.Name(), "help") {
			continue
		}
		ns = append(ns, c.Name())
	}
	return ns
}

func createFlags(cmd *cobra.Command) *Section {
	s := &Section{}
	s.Reset = resetForm(form.GetFormItemCount())

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		createFlagInput(cmd, flag)
	})
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		createFlagInput(cmd, flag)
	})

	form.AddInputField(argsLabel, "", 40, nil, func(text string) {
		updateOutput(form, cmd)
	})

	updateOutput(form, cmd)

	return s
}

func createFlagInput(cmd *cobra.Command, flag *pflag.Flag) {
	n := cases.Title(language.English).String(flag.Name)
	if form.GetFormItemByLabel(n) != nil {
		return
	}
	if n == "Output" {
		form.AddDropDown(n, outputs, 1, func(option string, optionIndex int) {
			updateOutput(form, cmd)
		})
		return
	}

	switch flag.Value.Type() {
	case "bool":
		form.AddCheckbox(n, internal.Must(strconv.ParseBool(flag.Value.String())), func(checked bool) {
			updateOutput(form, cmd)
		})
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		form.AddInputField(n, "", 40, tview.InputFieldInteger,
			func(text string) { updateOutput(form, cmd) })
	case "string":
		form.AddInputField(n, "", 40, nil, func(text string) {
			updateOutput(form, cmd)
		})
	case "stringArray", "stringSlice":
		form.AddInputField(n, "", 40, nil, func(text string) {
			updateOutput(form, cmd)
		})
	case "duration":
		form.AddInputField(n, "", 40, nil, func(text string) {
			updateOutput(form, cmd)
		})
	default:
		panic("unknown type: " + flag.Value.Type())
	}
}

func updateOutput(form *tview.Form, cmd *cobra.Command) {
	var flags []string
	for ; cmd != nil; cmd = cmd.Parent() {
		flags = append(flags, cmd.Name())
	}
	slices.Reverse(flags)

	for i := 0; i < form.GetFormItemCount(); i++ {
		it := form.GetFormItem(i)
		if strings.HasSuffix(it.GetLabel(), "Command") {
			continue
		}
		l := it.GetLabel()
		if l == argsLabel {
			flags = append(flags, highlight(it.(*tview.InputField).GetText()))
			continue
		}

		switch it.(type) {
		case *tview.InputField:
			val := it.(*tview.InputField).GetText()
			if val != "" {
				flags = append(flags, highlight("--"+(strings.ToLower(l)+"="+val)))
			}
		case *tview.Checkbox:
			if it.(*tview.Checkbox).IsChecked() {
				flags = append(flags, highlight("--"+(strings.ToLower(l))))
			}
		case *tview.DropDown:
			if _, opt := it.(*tview.DropDown).GetCurrentOption(); opt != "" &&
				(l != "Output" || opt != "json") {
				flags = append(flags, highlight("--"+(strings.ToLower(l)+"="+opt)))
			}
		}
	}

	line := strings.Join(flags, " ")
	cmdLine.SetText("")
	internal.Must(cmdLineWriter.Write([]byte(line)))
}

func highlight(str string) string {
	flag, val, ok := strings.Cut(str, "=")
	if !ok {
		return str
	}

	w := &strings.Builder{}
	internal.MustNoErr(quick.Highlight(w, val, b.Bash.Config().Name, "terminal", styles.Get("GitHub").Name))
	val = w.String()
	return flag + "=" + val
}

func execSimple(cmd *cobra.Command) string {
	console.Output = &strings.Builder{}
	defer func() { console.Output = os.Stdout }()
	if err := cmd.Execute(); err != nil {
		_, _ = console.Output.Write([]byte(err.Error()))
	}
	return console.Output.(*strings.Builder).String()
}

func execANSI(cmd *cobra.Command) string {
	stdout, stderr := os.Stdout, os.Stderr
	r, w, err := os.Pipe()
	internal.MustNoErr(err)
	defer func() { _, os.Stdout, os.Stderr, console.Output = r.Close(), stdout, stderr, stdout }()
	os.Stdout, os.Stderr = w, w
	console.Output = w

	if err = cmd.Execute(); err != nil {
		_, _ = console.Output.Write([]byte(err.Error()))
	}
	internal.MustNoErr(w.Close())
	bs, _ := io.ReadAll(r)
	return string(bs)
}

type Section struct {
	FormItem tview.FormItem
	PrevVal  string
	Child    *Section
	Reset    func()
}

func (s *Section) DisposeChildren() {
	if s.Child != nil {
		s.Child.DisposeChildren()
		if s.Child.Reset != nil {
			s.Child.Reset()
		}
		s.Child = nil
	}
}
