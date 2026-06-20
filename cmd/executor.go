package cmd

import (
	"fmt"
	"os"
	"vine-lang/env"
	"vine-lang/runner"
	"vine-lang/verror"
)

func init() {
	env.SetExecuteCode(runner.ExecuteCode)
}

func executeVineFile(filepath string, wk env.Workspace) error {
	defer func() {
		if r := recover(); r != nil {
			handleError(r)
		}
	}()

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %v", filepath, err)
	}

	_, err = runner.ExecuteCode(filepath, string(bytes), wk)

	return err
}

func handleError(r any) {
	switch err := r.(type) {
	case verror.VError:
		fmt.Fprintln(os.Stderr, err.Error())
	case verror.ParseVError:
		fmt.Fprintln(os.Stderr, err.Error())
	case verror.InterpreterVError:
		fmt.Fprintln(os.Stderr, err.Error())
	case verror.LexerVError:
		fmt.Fprintln(os.Stderr, err.Error())
	default:
		fmt.Fprintln(os.Stderr, r)
	}
}
