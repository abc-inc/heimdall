// This file contains code of Task.
// Copyright 2024 The Heimdall authors, Andrey Nering
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

//go:build !no_run

package run

import (
	"context"
	stdlog "log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-task/task/v3"
	"github.com/go-task/task/v3/args"
	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/interpreter"
	"github.com/go-task/task/v3/interpreter/exprext"
	"github.com/go-task/task/v3/taskfile"
	"github.com/mattn/go-isatty"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"mvdan.cc/sh/v3/syntax"
)

const usage = `Usage: task [-ilfwvsd] [--init] [--list] [--force] [--watch] [--verbose] [--silent] [--dir] [--taskfile] [--dry] [--summary] [task...]

Runs the specified task(s). Falls back to the "default" task if no task name
was specified, or lists all tasks if an unknown task name was specified.

Example: 'task hello' with the following 'Taskfile.yml' file will generate an
'output.txt' file with the content "hello".

'''
version: '3'
tasks:
  hello:
    cmds:
      - echo "I am going to write a file named 'output.txt' now."
      - echo "hello" > output.txt
    generates:
      - output.txt
'''

Options:
`

func runTask() {
	interpreter.Register("eval", func() interpreter.Interpreter {
		return exprext.ExprInterpreter{}
	})

	// log.SetFlags(0)
	// log.SetOutput(os.Stderr)

	pflag.Usage = func() {
		stdlog.Print(usage)
		pflag.PrintDefaults()
	}

	var (
		versionFlag bool
		helpFlag    bool
		init        bool
		list        bool
		listAll     bool
		listJson    bool
		status      bool
		force       bool
		watch       bool
		verbose     bool
		silent      bool
		assumeYes   bool
		dry         bool
		summary     bool
		exitCode    bool
		parallel    bool
		concurrency int
		dir         string
		entrypoint  string
		output      taskfile.Output
		color       bool
		interval    time.Duration
	)

	if pflag.Lookup("interval") == nil {
		// pflag.BoolVar(&versionFlag, "version", false, "show Task version")
		pflag.BoolVarP(&helpFlag, "help", "h", false, "shows Task usage")
		pflag.BoolVarP(&init, "init", "i", false, "creates a new Taskfile.yaml in the current folder")
		pflag.BoolVarP(&list, "list", "l", false, "lists tasks with description of current Taskfile")
		pflag.BoolVarP(&listAll, "list-all", "a", false, "lists tasks with or without a description")
		pflag.BoolVarP(&listJson, "json", "j", false, "formats task list as json")
		pflag.BoolVar(&status, "status", false, "exits with non-zero exit code if any of the given tasks is not up-to-date")
		pflag.BoolVarP(&force, "force", "f", false, "forces execution even when the task is up-to-date")
		pflag.BoolVarP(&watch, "watch", "w", false, "enables watch of the given task")
		pflag.BoolVarP(&verbose, "verbose", "v", false, "enables verbose mode")
		pflag.BoolVarP(&silent, "silent", "s", false, "disables echoing")
		pflag.BoolVarP(&assumeYes, "yes", "y", !isatty.IsTerminal(os.Stdin.Fd()), `Assume "yes" as answer to all prompts.`)
		pflag.BoolVarP(&parallel, "parallel", "p", false, "executes tasks provided on command line in parallel")
		pflag.BoolVarP(&dry, "dry", "n", false, "compiles and prints tasks in the order that they would be run, without executing them")
		pflag.BoolVar(&summary, "summary", false, "show summary about a task")
		pflag.BoolVarP(&exitCode, "exit-code", "x", false, "pass-through the exit code of the task command")
		pflag.StringVarP(&dir, "dir", "d", "", "sets directory of execution")
		pflag.StringVarP(&entrypoint, "taskfile", "t", "", `choose which Taskfile to run. Defaults to "Taskfile.yml"`)
		// pflag.StringVarP(&output.Name, "output", "o", "", "sets output style: [interleaved|group|prefixed]")
		pflag.StringVar(&output.Group.Begin, "output-group-begin", "", "message template to print before a task's grouped output")
		pflag.StringVar(&output.Group.End, "output-group-end", "", "message template to print after a task's grouped output")
		pflag.BoolVarP(&color, "color", "c", true, "colored output. Enabled by default. Set flag to false or use NO_COLOR=1 to disable")
		pflag.IntVarP(&concurrency, "concurrency", "C", 0, "limit number tasks to run concurrently")
		pflag.DurationVarP(&interval, "interval", "I", 0, "interval to watch for changes")
	}
	pflag.Parse()

	if versionFlag {
		// fmt.Printf("Task version: %s\n", ver.GetVersion())
		return
	}

	if helpFlag {
		pflag.Usage()
		return
	}

	if init {
		wd, err := os.Getwd()
		if err != nil {
			zlog.Fatal().Err(err).Send()
		}
		if err := task.InitTaskfile(os.Stdout, wd); err != nil {
			zlog.Fatal().Err(err).Send()
		}
		return
	}

	if dir != "" && entrypoint != "" {
		zlog.Fatal().Msg("task: You can't set both --dir and --taskfile")
		return
	}
	if entrypoint != "" {
		dir = filepath.Dir(entrypoint)
		entrypoint = filepath.Base(entrypoint)
	}

	if output.Name != "group" {
		if output.Group.Begin != "" {
			zlog.Fatal().Msg("task: You can't set --output-group-begin without --output=group")
			return
		}
		if output.Group.End != "" {
			zlog.Fatal().Msg("task: You can't set --output-group-end without --output=group")
			return
		}
	}

	e := task.Executor{
		Force:       force,
		Insecure:    true,
		Watch:       watch,
		Verbose:     verbose,
		Silent:      silent,
		AssumeYes:   assumeYes,
		Dir:         dir,
		Dry:         dry,
		Entrypoint:  entrypoint,
		Summary:     summary,
		Parallel:    parallel,
		Color:       color,
		Concurrency: concurrency,
		Interval:    interval,

		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,

		OutputStyle: output,
	}

	var listOptions = task.NewListOptions(list, listAll, listJson)
	if err := listOptions.Validate(); err != nil {
		zlog.Fatal().Err(err).Send()
	}

	if (listOptions.ShouldListTasks()) && silent {
		e.ListTaskNames(listAll)
		return
	}

	if err := e.Setup(); err != nil {
		zlog.Fatal().Err(err).Send()
	}
	v := e.Taskfile.Version
	if listOptions.ShouldListTasks() {
		if foundTasks, err := e.ListTasks(listOptions); !foundTasks || err != nil {
			os.Exit(1)
		}
		return
	}

	var (
		calls   []taskfile.Call
		globals *taskfile.Vars
	)

	tasksAndVars, cliArgs, err := getArgs()
	if err != nil {
		zlog.Fatal().Err(err).Send()
	}

	if v.Major() >= 3.0 {
		calls, globals = args.ParseV3(tasksAndVars...)
	} else {
		calls, globals = args.ParseV2(tasksAndVars...)
	}

	globals.Set("CLI_ARGS", taskfile.Var{Static: cliArgs})
	e.Taskfile.Vars.Merge(globals)

	if !watch {
		e.InterceptInterruptSignals()
	}

	ctx := context.Background()

	if status {
		if err := e.Status(ctx, calls...); err != nil {
			zlog.Fatal().Err(err).Send()
		}
		return
	}

	if err := e.Run(ctx, calls...); err != nil {
		// e.Logger.Errf(logger.Red, "%v", err)
		zlog.Err(err).Send()

		if exitCode {
			var err *errors.TaskRunError
			if errors.As(err, &err) {
				os.Exit(err.TaskExitCode())
			}
		}
		os.Exit(1)
	}
}

func getArgs() ([]string, string, error) {
	var (
		args          = pflag.Args()[1:]
		doubleDashPos = pflag.CommandLine.ArgsLenAtDash()
	)

	if doubleDashPos == -1 {
		return args, "", nil
	}

	var quotedCliArgs []string
	for _, arg := range args[doubleDashPos:] {
		quotedCliArg, err := syntax.Quote(arg, syntax.LangBash)
		if err != nil {
			return nil, "", err
		}
		quotedCliArgs = append(quotedCliArgs, quotedCliArg)
	}
	return args[:doubleDashPos], strings.Join(quotedCliArgs, " "), nil
}
