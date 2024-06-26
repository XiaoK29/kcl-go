package format

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCode(t *testing.T) {
	tcases := [...]struct {
		source    string
		expect    string
		expectErr string
	}{
		{
			source: "a=a+1",
			expect: "a = a + 1\n",
		},
		{
			source: "a=a+",
			expect: "a = a + \n",
		},
		{
			source: "a={a = 1,b=2}",
			expect: "a = {a = 1, b = 2}\n",
		},
	}

	for _, testCase := range tcases {
		actual, err := FormatCode(testCase.source)
		if testCase.expectErr != "" {
			assert.NotNil(t, err, "format code expect err, get no error")
			assert.Equal(t, testCase.expectErr, err.Error(), fmt.Sprintf("format code get wrong error result. expect: %s got: %s", testCase.expectErr, err.Error()))
		} else {
			assert.Equal(t, testCase.expect, string(actual), fmt.Sprintf("format file get wrong result. expect: %s got: %s", actual, testCase.expect))
		}
	}
}

func TestFormatPath(t *testing.T) {
	successDir := filepath.Join("testdata", "success")
	expectedFileSuffix := ".formatted"
	expectedFiles := findFiles(t, successDir, func(info fs.DirEntry) bool {
		return strings.HasSuffix(info.Name(), expectedFileSuffix)
	})

	sourceFiles := findFiles(t, successDir, func(info fs.DirEntry) bool {
		return strings.HasSuffix(info.Name(), ".k")
	})
	var sourceFilesBackup []kclFile

	for _, sourceFile := range sourceFiles {
		content, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("read source file content failed: %s", sourceFile)
		}
		sourceFilesBackup = append(sourceFilesBackup, kclFile{
			name:    sourceFile,
			content: content,
		})
	}

	changedPaths, err := FormatPath(successDir)
	// write back un-formatted file content
	defer writeFile(t, sourceFilesBackup)

	if err != nil {
		t.Fatalf("format path exec failed. %v", err)
	}

	var changedPathsRelative []string
	for _, p := range changedPaths {
		changedPathsRelative = append(changedPathsRelative, strings.TrimSuffix(p, ".k")+expectedFileSuffix)
	}
	assert.ElementsMatchf(t, expectedFiles, changedPathsRelative, "format path get wrong result. changedPath mismatch, expect: %s, get: %s", expectedFiles, changedPathsRelative)

	for _, expectedFile := range expectedFiles {
		expected, err := os.ReadFile(expectedFile)
		if runtime.GOOS == "windows" {
			expected = bytes.Replace(expected, []byte{0xd, 0xa}, []byte{0xa}, -1)
		}
		if err != nil {
			t.Fatalf("read expected formatted file failed: %s", expectedFile)
		}
		actualFile := strings.TrimSuffix(expectedFile, expectedFileSuffix) + ".k"
		get, err := os.ReadFile(actualFile)
		if err != nil {
			t.Fatalf("read actual formatted file failed: %s", actualFile)
		}
		assert.Equal(t, expected, get, fmt.Sprintf("format path get wrong result. formatted content mismatch, file: %s, expect: %s, get: %s", actualFile, expected, get))
	}
}

type filterFile func(fs.DirEntry) bool

func findFiles(t testing.TB, testDir string, filter filterFile) (names []string) {
	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, f := range files {
		if !f.IsDir() {
			if filter(f) {
				names = append(names, filepath.Join(testDir, f.Name()))
			}
		}
	}
	return names
}

type kclFile struct {
	name    string
	content []byte
}

func writeFile(t *testing.T, kclfiles []kclFile) {
	for _, backUpFile := range kclfiles {
		err := os.WriteFile(backUpFile.name, backUpFile.content, 0666)
		if err != nil {
			t.Logf("write back formatted source file failed: %v", err)
		}
	}
}
