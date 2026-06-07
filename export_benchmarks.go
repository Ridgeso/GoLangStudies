package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	registryFile = flag.String("registry", "./benchmarks.json", "input benchmark registry JSON")
	srcRoot      = flag.String("src", "./tests", "root directory that src_path values are relative to")
	outFile      = flag.String("out", "./results/presentation_data.json", "output JSON path")
)


type BenchmarkDef struct {
	Name        string `json:"name"`
	SrcPath     string `json:"src_path"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Internal    bool   `json:"internal"`
}

type SourceFile struct {
	RelPath  string `json:"rel_path"`
	Filename string `json:"filename"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

type BenchmarkExport struct {
	BenchmarkDef
	Files []SourceFile `json:"files"`
}

type ExportPayload struct {
	GeneratedAt int64             `json:"generated_at"`
	SrcRoot     string            `json:"src_root"`
	Benchmarks  []BenchmarkExport `json:"benchmarks"`
}

var extToLang = map[string]string{
	".go":  "go",
	".cpp": "cpp",
	".py":  "python",
}

var langOrder = map[string]int{"go": 0, "cpp": 1, "python": 2}

func collectFiles(dir, relBase string) ([]SourceFile, error) {
	var files []SourceFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		lang, ok := extToLang[strings.ToLower(filepath.Ext(path))]
		if !ok {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "  warn: could not read %s: %v\n", path, readErr)
			return nil
		}
		rel, _ := filepath.Rel(relBase, path)
		files = append(files, SourceFile{
			RelPath:  filepath.ToSlash(rel),
			Filename: info.Name(),
			Language: lang,
			Content:  string(content),
		})
		return nil
	})

	sort.Slice(files, func(i, j int) bool {
		oi := langOrder[files[i].Language]
		oj := langOrder[files[j].Language]
		if oi != oj {
			return oi < oj
		}
		return files[i].Filename < files[j].Filename
	})

	return files, err
}

func statusIcon(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

func main() {
	flag.Parse()

	regData, err := os.ReadFile(*registryFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot read registry %q: %v\n", *registryFile, err)
		os.Exit(1)
	}

	var registry []BenchmarkDef
	if err := json.Unmarshal(regData, &registry); err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid JSON in %q: %v\n", *registryFile, err)
		os.Exit(1)
	}

	fmt.Printf("registry : %s  (%d benchmarks)\n", *registryFile, len(registry))

	absRoot, err := filepath.Abs(*srcRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: bad --src path: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("src root : %s\n\n", absRoot)

	payload := ExportPayload{
		GeneratedAt: time.Now().Unix(),
		SrcRoot:     *srcRoot,
	}

	for _, def := range registry {
		dir := filepath.Join(absRoot, filepath.FromSlash(def.SrcPath))

		files, walkErr := collectFiles(dir, absRoot)
		if walkErr != nil {
			fmt.Fprintf(os.Stderr, "  warn: %s: %v\n", def.Name, walkErr)
		}

		seen := map[string]bool{}
		var langs []string
		for _, f := range files {
			if !seen[f.Language] {
				langs = append(langs, f.Language)
				seen[f.Language] = true
			}
		}

		if len(files) == 0 {
			fmt.Printf(
				"  %-36s  %s  (no files — expected at %s)\n",
				def.Name,
				statusIcon(false),
				filepath.Join(*srcRoot, def.SrcPath),
			)
		} else {
			fmt.Printf(
				"  %-36s  %s  %d file(s)  [%s]\n",
				def.Name,
				statusIcon(true),
				len(files),
				strings.Join(langs, ", "),
			)
		}

		payload.Benchmarks = append(payload.Benchmarks, BenchmarkExport{
			BenchmarkDef: def,
			Files:        files,
		})
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: marshal:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*outFile, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "error: write:", err)
		os.Exit(1)
	}

	totalFiles := 0
	for _, b := range payload.Benchmarks {
		totalFiles += len(b.Files)
	}
	fmt.Printf(
		"\n  wrote %d benchmarks, %d source files  →  %s\n",
		len(payload.Benchmarks), totalFiles, *outFile,
	)
}
