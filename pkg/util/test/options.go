// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"strings"
)

// Flag is a command line flag that has a key and a value.
type Flag struct {
	Key   string
	Value interface{}
}

// Slice returns a representation of this Flag as a slice of strings.
func (f *Flag) Slice() []string {
	key := fmt.Sprintf("--%s", f.Key)
	if f.Value == nil {
		return []string{key}
	}
	return []string{key, fmt.Sprintf("%v", f.Value)}
}

// String returns a representation of this Flag as a string.
func (f *Flag) String() string {
	return strings.Join(f.Slice(), " ")
}

// Command is a command that has a name, a list of flags, and a list of arguments.
type Command struct {
	Name  string
	Flags []Flag
	Args  []string
}

// CommandBuilder is a builder for Command objects.
type CommandBuilder struct {
	command Command
}

// NewCommandBuilder creates and returns a new CommandBuilder with the given name.
func NewCommandBuilder(name string) *CommandBuilder {
	return &CommandBuilder{Command{Name: name}}
}

// Flags appends the given flags to this CommandBuilder.
func (c *CommandBuilder) Flags(flags ...Flag) *CommandBuilder {
	c.command.Flags = append(c.command.Flags, flags...)
	return c
}

// BoolFlag appends a boolean flag with the given key to this CommandBuilder.
func (c *CommandBuilder) BoolFlag(key string) *CommandBuilder {
	return c.Flag(key, "")
}

// Flag appends a flag with the given key and value to this CommandBuilder.
func (c *CommandBuilder) Flag(key string, value interface{}) *CommandBuilder {
	return c.Flags(Flag{key, value})
}

// Args appends the given arguments to this CommandBuilder.
func (c *CommandBuilder) Args(args ...string) *CommandBuilder {
	c.command.Args = append(c.command.Args, args...)
	return c
}

// Command returns the Command that has been built by this CommandBuilder.
func (c *CommandBuilder) Command() *Command {
	return &c.command
}

// Slice returns a representation of this Command as a slice of strings.
func (c *Command) Slice() []string {
	out := []string{c.Name}
	for _, flag := range c.Flags {
		out = append(out, flag.Slice()...)
	}
	out = append(out, c.Args...)
	return out
}

// String returns a representation of this Command as a string.
func (c *Command) String() string {
	return strings.Join(c.Slice(), " ")
}
