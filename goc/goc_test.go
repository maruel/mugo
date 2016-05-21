//
// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package goc

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("goc:")
}

func TestTranspileUnit(t *testing.T) {
	data := []struct {
		in, expected string
	}{
		{
			"const a = 1",
			"const int a = 1;\n",
		},
		{
			"// Hi\nconst a = 1",
			"// Hi\nconst int a = 1;\n",
		},
		{
			"var a = 1",
			"int a = 1;\n",
		},
		{
			"var a = \"hi\"",
			"const char * a = \"hi\";\n",
		},
		{
			"const a = \"hi\"",
			"const char * const a = \"hi\";\n",
		},
		{
			"var (\na = 1\nb = 2\n)",
			"int a = 1;\nint b = 2;\n",
		},
		{
			"const (\na = 1\nb = 2\n)",
			"const int a = 1;\nconst int b = 2;\n",
		},
		{
			"func a() {}",
			"void a() {\n}\n",
		},
		{
			"// Import comment\nimport \"os\"",
			"// Import comment\n",
		},
		{
			"var a int",
			"int a = 0;\n",
		},
		{
			"var a string",
			"const char * a = \"\";\n",
		},
		{
			"var a, b int",
			"int a = 0;\nint b = 0;\n",
		},
		{
			"func a() {\n  a := 0\n}",
			"void a() {\n  int a = 0;\n}\n",
		},

		// These are incorrect:
		{
			"func a(){a:=0}",
			//"void a() {int a=0;}",
			"void a() {\n  int a = 0;\n}\n",
		},
	}
	b := &bytes.Buffer{}
	for i, line := range data {
		b.Reset()
		//log.Printf("%d: %q", i, line.in)
		in := "package a\n" + line.in
		if _, err := Transpile(b, strings.NewReader(in)); err != nil {
			t.Fatalf("%d: %s", i, err)
		}
		actual := b.String()
		if line.expected != actual {
			t.Fatalf("%d: Transpile(%v)\nExpected:\n%s\nActual:\n%s\n", i, in, line.expected, actual)
		}
	}
}

func TestTranspileDoc(t *testing.T) {
	// Test separately to reduce "package a\n" redundant string in TestTranspile.
	data := []struct {
		in, expected string
	}{
		{
			"package a",
			"",
		},
		{
			"package a\n",
			"",
		},
		{
			"// Hi\npackage a\nvar a = 1",
			"// Hi\nint a = 1;\n",
		},
		{
			"/* Hi */\npackage a\nvar a = 1",
			"/* Hi */\nint a = 1;\n",
		},

		// These are incorrect:
		{
			"// Hi\n\npackage a\nvar a = 1",
			"// Hi\nint a = 1;\n",
			//"// Hi\n\nint a = 1;\n",
		},
		{
			"// Hi\n\n// Hi2\n\npackage a\nvar a = 1",
			"// Hi\n// Hi2\nint a = 1;\n",
			//"// Hi\n\n// Hi2\n\nint a = 1;\n",
		},
	}
	b := &bytes.Buffer{}
	for i, line := range data {
		b.Reset()
		in := line.in
		if _, err := Transpile(b, strings.NewReader(in)); err != nil {
			t.Fatalf("%d: %s", i, err)
		}
		actual := b.String()
		if line.expected != actual {
			t.Fatalf("%d: Transpile(%v)\nExpected:\n%s\nActual:\n%s\n", i, in, line.expected, actual)
		}
	}
}

func TestTranspileError1(t *testing.T) {
	// Use a form that is known to be not supported, assert that line number is
	// properly calculated.
	in := strings.NewReader("// Comment\npackage a\n\nfunc a() (int, error) {}")
	out := &bytes.Buffer{}
	_, err := Transpile(out, in)
	//if out.String() != "// Comment\n" {
	if out.String() != "" {
		t.Fatalf("unexpected output: %q", out.String())
	}
	if err == nil {
		t.Fatal("expected error")
	}
	l := strings.SplitN(err.Error(), "\n", 2)
	if l[0] != "line 4: unsupported return type: &ast.FieldList{" {
		t.Fatalf("unexpected error: %q", l[0])
	}
}

func TestTranspileError2(t *testing.T) {
	out := &bytes.Buffer{}
	_, err := Transpile(out, &bytes.Buffer{})
	if out.String() != "" {
		t.Fatalf("unexpected output: %q", out.String())
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "failed to parse: src.go:1:1: expected 'package', found 'EOF'" {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}
