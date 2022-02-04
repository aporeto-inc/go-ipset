/*
Copyright 2015 Jan Broer All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package ipset is a library providing a wrapper to the IPtables ipset userspace utility
package ipset

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

type fakeCmdInterface interface {
	// CombinedOutput fakes the (*os.exec.Cmd).CombinedOutput()
	CombinedOutput() ([]byte, error)
}

type fakeCmd struct {
	cmd        string
	args       []string
	stdout     string
	stderr     string
	exitStatus int
	pathError  bool
}

// global for use with each command we execeute
var mockCommandEnv = []string{}
var mockCommandPathError = false

func (c *fakeCmd) setupMockCommandEnv() {
	env := []string{
		"GO_WANT_HELPER_PROCESS=1",
		"GO_WANT_FAILED_PATH=" + strconv.FormatBool(c.pathError),
		"GO_EXEC_HELPER_STDOUT=" + c.stdout,
		"GO_EXEC_HELPER_STDERR=" + c.stderr,
		"GO_EXEC_HELPER_EXIT_STATUS=" + strconv.Itoa(c.exitStatus),
	}
	mockCommandEnv = env
	mockCommandPathError = c.pathError
}

func mockExecCommand(command string, args ...string) *exec.Cmd {

	// want to force os.PathError
	if mockCommandPathError {
		return exec.Command("/blah/blah")
	}

	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = mockCommandEnv // use the global env
	return cmd
}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	t.Log("inside TestExecCommandHelper")
	t.Logf("GO_EXEC_HELPER_STDOUT='%s'", os.Getenv("GO_EXEC_HELPER_STDOUT"))
	t.Logf("GO_EXEC_HELPER_STDERR='%s'", os.Getenv("GO_EXEC_HELPER_STDERR"))
	t.Logf("GO_EXEC_HELPER_EXIT_STATUS='%s'", os.Getenv("GO_EXEC_HELPER_EXIT_STATUS"))

	fmt.Fprintf(os.Stdout, os.Getenv("GO_EXEC_HELPER_STDOUT"))
	fmt.Fprintf(os.Stderr, os.Getenv("GO_EXEC_HELPER_STDERR"))
	i, _ := strconv.Atoi(os.Getenv("GO_EXEC_HELPER_EXIT_STATUS"))
	os.Exit(i)
}

func Test_getIpsetVersionString(t *testing.T) {

	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	tests := []struct {
		name    string
		fakeCmd fakeCmd
		want    string
		wantErr bool
	}{
		{
			name: "test exit 0 stdout with vX.Y",
			want: "v7.6",
			fakeCmd: fakeCmd{
				cmd:        "ipset",
				args:       []string{"--version"},
				exitStatus: 0,
				stderr:     "",
				stdout:     "ipset v7.6, protocol version: 7",
			},
			wantErr: false,
		},
		{
			name: "test exit 0 stderr with vX.Y",
			want: "v2.3",
			fakeCmd: fakeCmd{
				cmd:        "ipset",
				args:       []string{"--version"},
				exitStatus: 0,
				stderr:     "ipset v2.3, protocol version: 2",
				stdout:     "",
			},
			wantErr: false,
		},
		{
			name: "test exit 0 stdout with vX.Y.Z",
			want: "v2.3",
			fakeCmd: fakeCmd{
				cmd:        "ipset",
				args:       []string{"--version"},
				exitStatus: 0,
				stderr:     "",
				stdout:     "ipset v2.3.4, protocol version: 7.2.3",
			},
			wantErr: false,
		},
		{
			name: "test exit 0 stdout with invalid version",
			want: "",
			fakeCmd: fakeCmd{
				cmd:        "ipset",
				args:       []string{"--version"},
				exitStatus: 0,
				stderr:     "",
				stdout:     "ipset blahblah, protocol? what protocol?",
			},
			wantErr: true,
		},
		{
			name: "test exit 1 stdout with good version string",
			want: "v7.5",
			fakeCmd: fakeCmd{
				cmd:        "ipset",
				args:       []string{"--version"},
				exitStatus: 1,
				stderr:     "",
				stdout:     "ipset v7.5, protocol version: 7",
			},
			wantErr: false,
		},
		{
			name: "test exit 1 with path not found error",
			want: "",
			fakeCmd: fakeCmd{
				cmd:        "ipset",
				args:       []string{"--version"},
				exitStatus: 0,
				stderr:     "",
				stdout:     "",
				pathError:  true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fakeCmd.setupMockCommandEnv()
			got, err := getIpsetVersionString()
			if (err != nil) != tt.wantErr {
				t.Errorf("getIpsetVersionString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getIpsetVersionString() = %v, want %v", got, tt.want)
			}
		})
	}
}
