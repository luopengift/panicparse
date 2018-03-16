// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"testing"

	"github.com/maruel/panicparse/stack"
)

var testPalette = &Palette{
	EOLReset:               "A",
	RoutineFirst:           "B",
	Routine:                "C",
	CreatedBy:              "D",
	Package:                "E",
	SourceFile:             "F",
	FunctionStdLib:         "G",
	FunctionStdLibExported: "H",
	FunctionMain:           "I",
	FunctionOther:          "J",
	FunctionOtherExported:  "K",
	Arguments:              "L",
}

func TestCalcLengths(t *testing.T) {
	b := stack.Buckets{
		{
			stack.Signature{
				Stack: stack.Stack{
					Calls: []stack.Call{
						{
							SourcePath: "/gopath/foo/baz.go",
							Line:       123,
							Func:       stack.Function{Raw: "main.func·001"},
						},
					},
				},
			},
			nil,
		},
	}
	srcLen, pkgLen := CalcLengths(b, true)
	// When printing, it prints the remote path, not the transposed local path.
	compareString(t, "/gopath/foo/baz.go:123", b[0].Signature.Stack.Calls[0].FullSourceLine())
	compareInt(t, len("/gopath/foo/baz.go:123"), srcLen)
	compareInt(t, len("foo"), pkgLen)

	srcLen, pkgLen = CalcLengths(b, false)
	compareString(t, "baz.go:123", b[0].Signature.Stack.Calls[0].SourceLine())
	compareInt(t, len("baz.go:123"), srcLen)
	compareInt(t, len("foo"), pkgLen)
}

func TestBucketHeader(t *testing.T) {
	b := &stack.Bucket{
		stack.Signature{
			State: "chan receive",
			CreatedBy: stack.Call{
				SourcePath: "/gopath/src/github.com/foo/bar/baz.go",
				Line:       74,
				Func:       stack.Function{Raw: "main.mainImpl"},
			},
			SleepMax: 6,
			SleepMin: 2,
		},
		[]stack.Goroutine{
			{
				First: true,
			},
			{},
		},
	}
	// When printing, it prints the remote path, not the transposed local path.
	compareString(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ /gopath/src/github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(b, true, true))
	compareString(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ /gopath/src/github.com/foo/bar/baz.go:74]A\n", testPalette.BucketHeader(b, true, false))
	compareString(t, "B2: chan receive [2~6 minutes]D [Created by main.mainImpl @ baz.go:74]A\n", testPalette.BucketHeader(b, false, true))
	compareString(t, "C2: chan receive [2~6 minutes]D [Created by main.mainImpl @ baz.go:74]A\n", testPalette.BucketHeader(b, false, false))

	b = &stack.Bucket{
		stack.Signature{
			State:    "b0rked",
			SleepMax: 6,
			SleepMin: 6,
			Locked:   true,
		},
		nil,
	}
	compareString(t, "C0: b0rked [6 minutes] [locked]A\n", testPalette.BucketHeader(b, false, false))
}

func TestStackLines(t *testing.T) {
	s := &stack.Signature{
		State: "idle",
		Stack: stack.Stack{
			Calls: []stack.Call{
				{
					SourcePath: "/goroot/src/runtime/sys_linux_amd64.s",
					Line:       400,
					Func:       stack.Function{Raw: "runtime.Epollwait"},
					Args: stack.Args{
						Values: []stack.Arg{
							{Value: 0x4},
							{Value: 0x7fff671c7118},
							{Value: 0xffffffff00000080},
							{},
							{Value: 0xffffffff0028c1be},
							{},
							{},
							{},
							{},
							{},
						},
						Elided: true,
					},
					IsStdlib: true,
				},
				{
					SourcePath: "/goroot/src/runtime/netpoll_epoll.go",
					Line:       68,
					Func:       stack.Function{Raw: "runtime.netpoll"},
					Args:       stack.Args{Values: []stack.Arg{{Value: 0x901b01}, {}}},
					IsStdlib:   true,
				},
				{
					SourcePath: "/gopath/src/main.go",
					Line:       1472,
					Func:       stack.Function{Raw: "main.Main"},
					Args:       stack.Args{Values: []stack.Arg{{Value: 0xc208012000}}},
				},
				{
					SourcePath: "/gopath/src/foo/bar.go",
					Line:       1575,
					Func:       stack.Function{Raw: "foo.OtherExported"},
				},
				{
					SourcePath: "/gopath/src/foo/bar.go",
					Line:       10,
					Func:       stack.Function{Raw: "foo.otherPrivate"},
				},
			},
			Elided: true,
		},
	}
	// When printing, it prints the remote path, not the transposed local path.
	expected := "" +
		"    Eruntime    F/goroot/src/runtime/sys_linux_amd64.s:400 HEpollwaitL(0x4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, 0, 0, 0, 0, 0, ...)A\n" +
		"    Eruntime    F/goroot/src/runtime/netpoll_epoll.go:68 GnetpollL(0x901b01, 0)A\n" +
		"    Emain       F/gopath/src/main.go:1472 IMainL(0xc208012000)A\n" +
		"    Efoo        F/gopath/src/foo/bar.go:1575 KOtherExportedL()A\n" +
		"    Efoo        F/gopath/src/foo/bar.go:10 JotherPrivateL()A\n" +
		"    (...)\n"
	compareString(t, expected, testPalette.StackLines(s, 10, 10, true))
	expected = "" +
		"    Eruntime    Fsys_linux_amd64.s:400 HEpollwaitL(0x4, 0x7fff671c7118, 0xffffffff00000080, 0, 0xffffffff0028c1be, 0, 0, 0, 0, 0, ...)A\n" +
		"    Eruntime    Fnetpoll_epoll.go:68 GnetpollL(0x901b01, 0)A\n" +
		"    Emain       Fmain.go:1472 IMainL(0xc208012000)A\n" +
		"    Efoo        Fbar.go:1575 KOtherExportedL()A\n" +
		"    Efoo        Fbar.go:10  JotherPrivateL()A\n" +
		"    (...)\n"
	compareString(t, expected, testPalette.StackLines(s, 10, 10, false))
}

func compareString(t *testing.T, expected, actual string) {
	if expected != actual {
		i := 0
		for i < len(expected) && i < len(actual) && expected[i] == actual[i] {
			i++
		}
		t.Fatalf("Delta at offset %d:\n- %q\n- %q", i, expected, actual)
	}
}

func compareInt(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Fatalf("%d != %d", expected, actual)
	}
}