//go:build cgo
// +build cgo

package native

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	_ "kcl-lang.io/kcl-go/pkg/plugin/hello_plugin"
	"kcl-lang.io/kcl-go/pkg/spec/gpyrpc"
)

const code = `
import kcl_plugin.hello

name = "kcl"
sum = hello.add(option("a"), option("b"))
`

func TestExecProgramWithPlugin(t *testing.T) {
	client := NewNativeServiceClient()
	result, err := client.ExecProgram(&gpyrpc.ExecProgram_Args{
		KFilenameList: []string{"main.k"},
		KCodeList:     []string{code},
		Args: []*gpyrpc.Argument{
			{
				Name:  "a",
				Value: "1",
			},
			{
				Name:  "b",
				Value: "2",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ErrMessage != "" {
		t.Fatal("error message must be empty")
	}
}

func TestExecProgramWithPluginError(t *testing.T) {
	client := NewNativeServiceClient()
	result, err := client.ExecProgram(&gpyrpc.ExecProgram_Args{
		KFilenameList: []string{"main.k"},
		KCodeList:     []string{code},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.ErrMessage, "strconv.ParseInt: parsing \"<nil>\": invalid syntax") {
		t.Fatal(result.ErrMessage)
	}
}

func TestParseFile(t *testing.T) {
	// Example: Test with string source
	src := `schema Name:
    name: str

n = Name {name = "name"}` // Sample KCL source code
	astJson, err := ParseFileASTJson("", src)
	if err != nil {
		t.Errorf("ParseFileASTJson failed with string source: %v", err)
	}
	if astJson == "" {
		t.Errorf("Expected non-empty AST JSON with string source")
	}

	// Example: Test with byte slice source
	srcBytes := []byte(src)
	astJson, err = ParseFileASTJson("", srcBytes)
	if err != nil {
		t.Errorf("ParseFileASTJson failed with byte slice source: %v", err)
	}
	if astJson == "" {
		t.Errorf("Expected non-empty AST JSON with byte slice source")
	}

	startTime := time.Now()
	// Example: Test with io.Reader source
	srcReader := strings.NewReader(src)
	astJson, err = ParseFileASTJson("", srcReader)
	if err != nil {
		t.Errorf("ParseFileASTJson failed with io.Reader source: %v", err)
	}
	if astJson == "" {
		t.Errorf("Expected non-empty AST JSON with io.Reader source")
	}
	elapsed := time.Since(startTime)
	t.Logf("ParseFileASTJson took %s", elapsed)
	if !strings.Contains(astJson, "Schema") {
		t.Errorf("Expected Schema Node AST JSON with io.Reader source")
	}
	if !strings.Contains(astJson, "Assign") {
		t.Errorf("Expected Assign Node AST JSON with io.Reader source")
	}
}

func ParseFileASTJson(filename string, src interface{}) (result string, err error) {
	var code string
	if src != nil {
		switch src := src.(type) {
		case []byte:
			code = string(src)
		case string:
			code = src
		case io.Reader:
			d, err := io.ReadAll(src)
			if err != nil {
				return "", err
			}
			code = string(d)
		default:
			return "", fmt.Errorf("unsupported src type: %T", src)
		}
	}
	client := NewNativeServiceClient()
	resp, err := client.ParseFile(&gpyrpc.ParseFile_Args{
		Path:   filename,
		Source: code,
	})
	if err != nil {
		return "", err
	}
	return resp.AstJson, nil
}
