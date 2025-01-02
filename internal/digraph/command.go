// Copyright (C) 2024 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package digraph

import (
	"fmt"
	"strings"

	"github.com/dagu-org/dagu/internal/cmdutil"
)

// buildCommand parses the command field in the step definition.
// Case 1: command is nil
// Case 2: command is a string
// Case 3: command is an array
//
// In case 3, the first element is the command and the rest are the arguments.
// If the arguments are not strings, they are converted to strings.
//
// Example:
// ```yaml
// step:
//   - name: "echo hello"
//     command: "echo hello"
//
// ```
// or
// ```yaml
// step:
//   - name: "echo hello"
//     command: ["echo", "hello"]
//
// ```
// It returns an error if the command is not nil but empty.
func buildCommand(_ BuildContext, def stepDef, step *Step) error {
	command := def.Command

	// Case 1: command is nil
	if command == nil {
		return nil
	}

	switch val := command.(type) {
	case string:
		// Case 2: command is a string
		if val == "" {
			return WrapError("command", val, errStepCommandIsEmpty)
		}
		// We need to split the command into command and args.
		step.CmdWithArgs = val
		cmd, args, err := cmdutil.SplitCommand(val)
		if err != nil {
			return WrapError("command", val, fmt.Errorf("failed to parse command: %w", err))
		}
		step.Command = cmd
		step.Args = args

	case []any:
		// Case 3: command is an array
		for _, v := range val {
			val, ok := v.(string)
			if !ok {
				// If the value is not a string, convert it to a string.
				// This is useful when the value is an integer for example.
				val = fmt.Sprintf("%v", v)
			}
			if step.Command == "" {
				step.Command = val
				continue
			}
			step.Args = append(step.Args, val)
		}

		// Setup CmdWithArgs (this will be actually used in the command execution)
		var sb strings.Builder
		for i, arg := range step.Args {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%q", arg))
		}
		step.CmdWithArgs = fmt.Sprintf("%s %s", step.Command, sb.String())

	default:
		// Unknown type for command field.
		return WrapError("command", val, errStepCommandMustBeArrayOrString)

	}

	return nil
}
