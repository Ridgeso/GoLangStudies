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
)

var (
	runs 		= flag.Int	 ("runs"        , 3    , "")
	timeout 	= flag.Int	 ("timeout"     , 180  , "")
	filter 		= flag.String("filter"      , ""   , "")
	noGo 		= flag.Bool  ("no-go"       , false, "")
	noCpp 		= flag.Bool  ("no-cpp"      , false, "")
	noPy 		= flag.Bool  ("no-python"   , false, "")
	output 		= flag.String("output"	     , ""   , "")
	cppCompiler = flag.String("cpp-compiler", "g++", "")
	goCompiler 	= flag.String("go-compiler" , "go" , "")
)

type RunResult struct {
	ExitCode   int     `json:"exit_code"`
	Stdout     string  `json:"stdout"`
	Stderr     string  `json:"stderr"`
	WallSecond float64 `json:"wall_seconds"`
}

type TestResult struct {
	TestName   string	 `json:"test_name"`
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
}

type BenchmarkResult struct {
	Name       string    	`json:"name"`
	Internal   bool		 	`json:"internal"`
	Category   string	 	`json:"category"`
	Results    []TestResult `json:"results"`
}

type Benchmark struct {
	Name        string
	SrcPath     string
	Category    string
	Description string
	Internal	bool
}

var benchmarks = []Benchmark{
	{"concurrency",    			  "concurrency",    		  "concurrency", "Concurrency benchmark", 		  false},
	{"fibonacci_iter", 			  "fibonacci",      		  "cpu", 		 "Fibonacci(40) iterative", 	  false},
	{"fibonacci_rec",  			  "fibRecursive",   		  "cpu", 		 "Fibonacci(40) recursive", 	  false},
	{"file_io",		   			  "fileIO", 				  "io",          "File IO", 					  false},
	{"hashmap",	       			  "hashmap",        		  "data",        "Hashmap benchmark", 			  false},
	{"json_roundtrip", 			  "jsonParser",     		  "data", 	     "JSON encode/decode", 			  false},
	{"matrix_mul",	   			  "matrixMultiply", 		  "cpu", 		 "400x400 matrix multiplication", false},
	{"prime_sieve",	   			  "primeSieve",     		  "cpu", 		 "Sieve of Eratosthenes", 		  false},
	{"sort_ints",	   			  "sorting", 	    		  "cpu", 		 "Sort integers", 				  false},
	{"string_build_multi_lang",   "stringBuilding/multiLang", "data",        "String building", 			  false},
	{"string_build_variants",     "stringBuilding/variants",  "data",        "Variations on string building", true},
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

func compileGo(src, out string, timeout int) (float64, string, bool) {
	r := runCmd(timeout, "", *goCompiler, "build", "-o", out, src)
	return r.WallSecond, r.Stderr, r.ExitCode == 0
}

func compileCpp(src, out string, timeout int) (float64, string, bool) {
	r := runCmd(timeout, "", *cppCompiler, "-O2", "-std=c++17", "-pthread", "-o", out, src)
	return r.WallSecond, r.Stderr, r.ExitCode == 0
}

func buildTestSources(testSrc string) ([]string, error) {
	var srcs []string
	srcBase := filepath.Join(SRC, testSrc)
	err := filepath.Walk(
		srcBase,
		func(path string, f os.FileInfo, err error) error {
			if filepath.Ext(path) == "" {
				return err
			}
			srcs = append(srcs, path)
			return err
		})
	return srcs, err
}

func skipLang(lang string) bool {
	switch lang {
		case "go":
			return *noGo
		case "c++":
			return *noCpp
		case "python":
			return *noPy
		default:
			return false
	}
}

func mapExt2Lang(path string) string {
	switch filepath.Ext(path) {
		case ".go":
			return "go"
		case ".cpp":
			return "c++"
		case ".py":
			return "python"
		default:
			return "idk"
	}
}

func runLangBenchmark(name, lang, src string, runs, timeout int) TestResult {	
	var binary string
	sum := 0.0

	testName := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
	res := TestResult{TestName: testName, Language: lang}

	if skipLang(lang) {
		res.Skipped = true
		res.SkipReason = "filtered lang: " + lang
		goto finishBenchmark
	}

	if _, err := os.Stat(src); err != nil {
		res.Skipped = true
		res.SkipReason = "source not found: " + src
		goto finishBenchmark
	}

	
	switch lang {
		case "go":
			binary = filepath.Join(BUILD_DIR, name+"_go")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			ct, ce, ok := compileGo(src, binary, timeout)
			res.CompileS = &ct
			if !ok {
				res.Skipped = true
				res.Error = ce
				res.SkipReason = ce
				return res
			}
		case "c++":
			binary = filepath.Join(BUILD_DIR, name+"_cpp")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			ct, ce, ok := compileCpp(src, binary, timeout)
			res.CompileS = &ct
			if !ok {
				res.Skipped = true
				res.Error = ce
				res.SkipReason = ce
				goto finishBenchmark
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
			goto finishBenchmark
		}

		res.RunsS = append(res.RunsS, r.WallSecond)
		if i == 0 {
			res.Output = r.Stdout
		}
	}

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

finishBenchmark:
	fmt.Printf("  %-7s skipped=%v exec=%.4fs\n", lang, res.Skipped, res.ExecMeanS)
	return res
}

func runBenchmark(b Benchmark, runs, timeout int) BenchmarkResult {	
	res := BenchmarkResult{
		Name: b.Name,
		Internal: b.Internal,
		Category: b.Category,
		Results: make([]TestResult, 0),
	}

	srcs, err := buildTestSources(b.SrcPath)
	if err != nil {
		fmt.Println("problem with reading ", b.Name," files: ", err)
		return res
	}

	for _, src := range srcs {
		lang := mapExt2Lang(src)
		r := runLangBenchmark(b.Name, lang, src, runs, timeout)
		res.Results = append(res.Results, r)
	}

	return res
}

func init() {
	flag.Parse()

	_ = os.MkdirAll(BUILD_DIR, 0755)
	_ = os.MkdirAll(RESULTS, 0755)
}

func main() {
	var results []BenchmarkResult
	start := time.Now()

	for _, b := range benchmarks {
		if *filter != "" && !strings.Contains(strings.ToLower(b.Name), strings.ToLower(*filter)) {
			continue
		}

		fmt.Printf("[%s] %s\n", strings.ToUpper(b.Category), b.Name)

		r := runBenchmark(b, *runs, *timeout)
		results = append(results, r)
	}

	payload := map[string]any{
		"timestamp":		  time.Now().Unix(),
		"runs_per_benchmark": *runs,
		"timeout_s":		  *timeout,
		"total_wall_s":		  time.Since(start).Seconds(),
		"benchmarks":		  results,
	}

	data, _ := json.MarshalIndent(payload, "", "  ")

	out := *output
	if out == "" {
		out = filepath.Join(RESULTS, fmt.Sprintf("run_%d.json", time.Now().Unix()))
	}

	_ = os.WriteFile(out, data, 0644)
	fmt.Println("Results written ->", out)
}
