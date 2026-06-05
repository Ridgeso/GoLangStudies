package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	ROOT, _   = os.Getwd()
	SRC       = filepath.Join(ROOT, "tests")
	RESULTS   = filepath.Join(ROOT, "results")
	BUILD_DIR = filepath.Join(ROOT, "build")

	GO_NAME  = "goSrc.go"
	CPP_NAME = "cppSrc.cpp"
	PY_NAME  = "pythonSrc.py"
)

type RunResult struct {
	ExitCode   int     `json:"exit_code"`
	Stdout     string  `json:"stdout"`
	Stderr     string  `json:"stderr"`
	WallSecond float64 `json:"wall_seconds"`
}

type BenchmarkResult struct {
	Name       string    `json:"name"`
	Language   string    `json:"language"`
	CompileS   *float64  `json:"compile_s"`
	RunsS      []float64 `json:"runs_s"`
	ExecMeanS  float64   `json:"exec_mean_s"`
	ExecMinS   float64   `json:"exec_min_s"`
	ExecMaxS   float64   `json:"exec_max_s"`
	TotalS     float64   `json:"total_s"`
	Output     string    `json:"output"`
	Error      string    `json:"error"`
	Skipped    bool      `json:"skipped"`
	SkipReason string    `json:"skip_reason"`
	Internal   bool		 `json:"internal"`
}

type Benchmark struct {
	Name        string
	SrcPath     string
	Category    string
	Description string
}

var benchmarks = []Benchmark{
	{"concurrency",    			  "concurrency",    		  "concurrency", "Concurrency benchmark"},
	{"fibonacci_iter", 			  "fibonacci",      		  "cpu", 		 "Fibonacci(40) iterative"},
	{"fibonacci_rec",  			  "fibRecursive",   		  "cpu", 		 "Fibonacci(42) recursive"},
	{"file_io",		   			  "fileIO", 				  "io",          "File IO"},
	{"hashmap",	       			  "hashmap",        		  "data",        "Hashmap benchmark"},
	{"json_roundtrip", 			  "jsonParser",     		  "data", 	     "JSON encode/decode"},
	{"matrix_mul",	   			  "matrixMultiply", 		  "cpu", 		 "400x400 matrix multiplication"},
	{"prime_sieve",	   			  "primeSieve",     		  "cpu", 		 "Sieve of Eratosthenes"},
	{"sort_ints",	   			  "sorting", 	    		  "cpu", 		 "Sort integers"},
	{"string_build_multi_lang",   "stringBuilding/multiLang", "data",        "String builder"},
}

func runCmd(timeout int, cwd string, cmd string, args ...string) RunResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, cmd, args...)
	if cwd != "" {
		c.Dir = cwd
	}

	out, err := c.Output()
	elapsed := time.Since(start).Seconds()

	if ctx.Err() == context.DeadlineExceeded {
		return RunResult{-1, "", fmt.Sprintf("TIMEOUT after %ds", timeout), elapsed}
	}

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return RunResult{ee.ExitCode(), string(out), string(ee.Stderr), elapsed}
		}
		return RunResult{-2, "", err.Error(), elapsed}
	}

	return RunResult{0, strings.TrimSpace(string(out)), "", elapsed}
}

func compileGo(src, out string, timeout int, goBin string) (float64, string, bool) {
	r := runCmd(timeout, "", goBin, "build", "-o", out, src)
	return r.WallSecond, r.Stderr, r.ExitCode == 0
}

func compileCpp(src, out string, timeout int, cpp string) (float64, string, bool) {
	r := runCmd(timeout, "", cpp, "-O2", "-std=c++17", "-pthread", "-o", out, src)
	return r.WallSecond, r.Stderr, r.ExitCode == 0
}

func runBenchmark(name, lang, src string, runs, timeout int, goBin, cppBin string) BenchmarkResult {
	res := BenchmarkResult{Name: name, Language: lang}

	if _, err := os.Stat(src); err != nil {
		res.Skipped = true
		res.SkipReason = "source not found: " + src
		return res
	}

	var binary string

	switch lang {
		case "go":
			binary = filepath.Join(BUILD_DIR, name+"_go")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			ct, ce, ok := compileGo(src, binary, timeout, goBin)
			res.CompileS = &ct
			if !ok {
				res.Skipped = true
				res.Error = ce
				res.SkipReason = ce
				return res
			}
		case "cpp":
			binary = filepath.Join(BUILD_DIR, name+"_cpp")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			ct, ce, ok := compileCpp(src, binary, timeout, cppBin)
			res.CompileS = &ct
			if !ok {
				res.Skipped = true
				res.Error = ce
				res.SkipReason = ce
				return res
			}
		default:
			v := 0.0
			res.CompileS = &v
	}

	for i := 0; i < runs; i++ {
		var r RunResult

		if lang == "python" {
			r = runCmd(timeout, "", "python", src)
		} else {
			r = runCmd(timeout, "", binary)
		}

		if r.ExitCode != 0 {
			res.Skipped = true
			res.Error = r.Stderr
			res.SkipReason = r.Stderr
			return res
		}

		res.RunsS = append(res.RunsS, r.WallSecond)
		if i == 0 {
			res.Output = r.Stdout
		}
	}

	sum := 0.0
	res.ExecMinS = math.MaxFloat64
	for _, t := range res.RunsS {
		sum += t
		if t < res.ExecMinS {
			res.ExecMinS = t
		}
		if t > res.ExecMaxS {
			res.ExecMaxS = t
		}
	}
	if len(res.RunsS) > 0 {
		res.ExecMeanS = sum / float64(len(res.RunsS))
	}
	res.TotalS = *res.CompileS + res.ExecMeanS
	return res
}

func main() {
	runs 		:= flag.Int	  ("runs"        , 3    , "")
	timeout 	:= flag.Int	  ("timeout"     , 180  , "")
	filter 		:= flag.String("filter"      , ""   , "")
	noGo 		:= flag.Bool  ("no-go"       , false, "")
	noCpp 		:= flag.Bool  ("no-cpp"      , false, "")
	noPy 		:= flag.Bool  ("no-python"   , false, "")
	output 		:= flag.String("output"	     , ""   , "")
	cppCompiler := flag.String("cpp-compiler", "g++", "")
	goCompiler 	:= flag.String("go-compiler" , "go" , "")
	flag.Parse()

	_ = os.MkdirAll(BUILD_DIR, 0755)
	_ = os.MkdirAll(RESULTS, 0755)

	var langs []string
	if !*noGo {
		langs = append(langs, "go")
	}
	if !*noCpp {
		langs = append(langs, "cpp")
	}
	if !*noPy {
		langs = append(langs, "python")
	}

	var results []BenchmarkResult
	start := time.Now()

	for _, b := range benchmarks {
		if *filter != "" && !strings.Contains(strings.ToLower(b.Name), strings.ToLower(*filter)) {
			continue
		}

		srcBase := filepath.Join(SRC, b.SrcPath)
		srcs := map[string]string{
			"go": filepath.Join(srcBase, GO_NAME),
			"cpp": filepath.Join(srcBase, CPP_NAME),
			"python": filepath.Join(srcBase, PY_NAME),
		}

		fmt.Printf("[%s] %s\n", strings.ToUpper(b.Category), b.Name)

		for _, lang := range langs {
			r := runBenchmark(b.Name, lang, srcs[lang], *runs, *timeout, *goCompiler, *cppCompiler)
			results = append(results, r)
			fmt.Printf("  %-7s skipped=%v exec=%.4fs\n", lang, r.Skipped, r.ExecMeanS)
		}
	}

	payload := map[string]any{
		"timestamp": time.Now().Unix(),
		"runs_per_benchmark": *runs,
		"timeout_s": *timeout,
		"total_wall_s": time.Since(start).Seconds(),
		"results": results,
	}

	data, _ := json.MarshalIndent(payload, "", "  ")

	out := *output
	if out == "" {
		out = filepath.Join(RESULTS, fmt.Sprintf("run_%d.json", time.Now().Unix()))
	}

	_ = os.WriteFile(out, data, 0644)
	fmt.Println("Results written ->", out)
}
