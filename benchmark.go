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
	"strconv"
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
	runs        = flag.Int   ("runs"        , 3    , "Number of runs per test")
	timeout     = flag.Int   ("timeout"     , 180  , "Time after test is being stoped")
	filter      = flag.String("filter"      , ""   , "Filter all tests maching this value")
	noGo        = flag.Bool  ("no-go"       , false, "Filter Go tests")
	noCpp       = flag.Bool  ("no-cpp"      , false, "Filter C++ tests")
	noPy        = flag.Bool  ("no-python"   , false, "Filter Python tests")
	output      = flag.String("output"      , ""   , "Set the output path")
	cppCompiler = flag.String("cpp-compiler", "g++", "Provide c++ bin path")
	goCompiler  = flag.String("go-compiler" , "go" , "Provide go bin path")
	noColor     = flag.Bool  ("no-color"    , false, "disable ANSI colors")
)

const (
    esc   = "\033["
    reset = esc + "0m"
    bold  = esc + "1m"
    dim   = esc + "2m"

    fgBlack   = esc + "30m"
    fgRed     = esc + "31m"
    fgGreen   = esc + "32m"
    fgYellow  = esc + "33m"
    fgBlue    = esc + "34m"
    fgMagenta = esc + "35m"
    fgCyan    = esc + "36m"
    fgWhite   = esc + "37m"

    fgBrightBlack   = esc + "90m"
    fgBrightRed     = esc + "91m"
    fgBrightGreen   = esc + "92m"
    fgBrightYellow  = esc + "93m"
    fgBrightBlue    = esc + "94m"
    fgBrightMagenta = esc + "95m"
    fgBrightCyan    = esc + "96m"
    fgBrightWhite   = esc + "97m"

    bgBlue = esc + "44m"

    clearLine = "\r" + esc + "2K"
)

var colorEnabled = true

func c(codes ...string) string {
	if !colorEnabled {
		return ""
	}
	return strings.Join(codes, "")
}

func colored(text string, codes ...string) string {
	if !colorEnabled {
		return text
	}
	return strings.Join(codes, "") + text + reset
}

func langColor(lang string) string {
	switch lang {
        case "go":
            return fgBrightCyan
        case "c++":
            return fgBrightBlue
        case "python":
            return fgBrightYellow
        default:
            return fgBrightMagenta
	}
}

func langTag(lang string) string {
	tag := fmt.Sprintf(" %-7s", lang)
	return colored(tag, bold, langColor(lang))
}

func categoryColor(cat string) string {
	switch cat {
        case "cpu":
            return fgBrightRed
        case "data":
            return fgBrightGreen
        case "io":
            return fgBrightYellow
        case "concurrency":
            return fgBrightMagenta
        default:
            return fgBrightWhite
	}
}

func fmtTime(s float64) string {
	var text string
	var col string
	switch {
        case s <= 0:
            return colored("  —      ", dim)
        case s < 0.001:
            text = fmt.Sprintf("%7.3fms", s*1000)
            col = fgBrightGreen
        case s < 0.1:
            text = fmt.Sprintf("%7.2fms", s*1000)
            col = fgBrightGreen
        case s < 1.0:
            text = fmt.Sprintf("%7.0fms", s*1000)
            col = fgBrightYellow
        case s < 10:
            text = fmt.Sprintf("%7.3f s", s)
            col = fgYellow
        default:
            text = fmt.Sprintf("%7.2f s", s)
            col = fgRed
	}
	return colored(text, col)
}

const barWidth = 28

func progressBar(done, total int, label string) string {
	pct := float64(done) / float64(total)
	filled := int(pct * barWidth)
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	bar := colored("[", dim) +
		colored(strings.Repeat("█", filled), fgBrightCyan) +
		colored(strings.Repeat("░", empty), fgBrightBlack) +
		colored("]", dim)

	pctStr := colored(fmt.Sprintf("%3.0f%%", pct*100), bold, fgBrightWhite)
	labelStr := colored(fmt.Sprintf(" %-28s", truncate(label, 28)), fgBrightBlack)

	return fmt.Sprintf("%s %s%s", bar, pctStr, labelStr)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func printProgress(done, total int, label string) {
	fmt.Print(clearLine + progressBar(done, total, label))
}

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func spinnerFrame(i int) string {
	return colored(spinFrames[i%len(spinFrames)], fgBrightCyan)
}

func printBenchHeader(idx, total int, b Benchmark) {
	catCol := categoryColor(b.Category)
	catTag := colored(fmt.Sprintf(" %-12s", strings.ToUpper(b.Category)), bold, catCol)
	nameStr := colored(b.Name, bold, fgBrightWhite)
	descStr := colored("  "+b.Description, dim)
	progress := colored(fmt.Sprintf("[%d/%d]", idx+1, total), fgBrightBlack)

	fmt.Printf("\n%s%s %s%s %s\n", catTag, reset, nameStr, descStr, progress)
	fmt.Println(colored(strings.Repeat("─", 64), fgBrightBlack))
}

func printResultRow(res TestResult) {
	tag := langTag(res.Language)

	if res.Skipped {
		reason := truncate(res.SkipReason, 48)
		fmt.Printf("  %s  %s %s\n", tag, colored("SKIPPED", bold, fgBrightBlack), colored(reason, dim))
		return
	}

	compileStr := ""
	if res.CompileS != nil && *res.CompileS > 0 {
		compileStr = colored("compile ", dim) + fmtTime(*res.CompileS) + "  "
	} else if res.CompileS != nil {
		compileStr = colored("compile   interp  ", dim) + "  "
	}

	execStr := colored("exec ", dim) +
		fmtTime(res.ExecMeanS) +
		colored(fmt.Sprintf("  (min %s  max %s)", fmtTime(res.ExecMinS), fmtTime(res.ExecMaxS)), dim)

	fmt.Printf("  %s  %s%s\n", tag, compileStr, execStr)
}

func printRunProgress(run, total int, lang string) {
	dots := strings.Repeat(colored("●", fgBrightCyan), run) + strings.Repeat(colored("○", fgBrightBlack), total-run)
	label := colored(fmt.Sprintf(" run %d/%d of %s", run, total, colored(lang, langColor(lang))), fgBrightBlack)
	fmt.Printf("%s  %s%s", clearLine, dots, label)
}

func printSummaryRow(br BenchmarkResult) {
    catCol := categoryColor(br.Category)
    catStr := colored(fmt.Sprintf("  %-12s", br.Category), catCol)

    marker := " "
    if br.Internal {
        marker = colored("⊙", fgBrightMagenta)
    }
    nameStr := fmt.Sprintf("%s %-28s", marker, truncate(br.Name, 27))

    times := map[string]float64{}
    for i, r := range br.Results {
        if !r.Skipped {
            lang := r.Language
            if br.Internal {
                lang += strconv.Itoa(i + 1)
            }
            times[lang] = r.ExecMeanS
        }
    }

    best := math.MaxFloat64
    for _, t := range times {
        if t < best {
            best = t
        }
    }

    fmtCell := func(lang string) string {
        t, ok := times[lang]
        if !ok {
            return colored("    —    ", fgBrightBlack)
        }
        s := fmtTime(t)
        if t == best {
            return colored("★", fgBrightYellow) + s
        }
        return " " + s
    }

    outFmt := ""
    if br.Internal {
        for i := range br.Results {
            outFmt += fmtCell("go" + strconv.Itoa(i + 1)) + "  "
        }
    } else {
        outFmt = fmt.Sprintf("%s  %s  %s", fmtCell("go"), fmtCell("c++"), fmtCell("python"))

    }

    fmt.Printf(
        "%s  %s  %s\n",
        catStr,
        nameStr,
        outFmt,
    )
}

func printSummary(results []BenchmarkResult, totalWall float64) {
	fmt.Printf("\n%s\n", colored(strings.Repeat("═", 72), fgBrightBlack))
	fmt.Printf("%s\n\n", colored("  RESULTS", bold, fgBrightWhite))

	fmt.Printf("%s\n", colored("  MULTILANG", bold, fgBrightWhite))
	fmt.Printf(
        "  %-15s    %-28s                   %s   %s  %s\n",
		colored("CATEGORY    ", bold, fgBrightBlack),
		colored("BENCHMARK", bold, fgBrightBlack),
		colored("    GO   ", bold, fgBrightCyan),
		colored("   C++   ", bold, fgBrightBlue),
		colored(" PYTHON  ", bold, fgBrightYellow),
	)
	fmt.Println(colored("  "+strings.Repeat("─", 68), fgBrightBlack))

	for _, br := range results {
        if !br.Internal {
            printSummaryRow(br)
        }
    }

	fmt.Printf("\n%s\n", colored("  INTERNAL", bold, fgBrightWhite))
    fmt.Printf(
        "  %-15s    %-28s                   %s\n",
		colored("CATEGORY    ", bold, fgBrightBlack),
		colored("BENCHMARK", bold, fgBrightBlack),
		colored("   GOs...", bold, fgBrightCyan),
	)
	fmt.Println(colored("  "+strings.Repeat("─", 68), fgBrightBlack))

	for _, br := range results {
        if br.Internal {
            printSummaryRow(br)
        }
    }

    fmt.Println(colored("  "+strings.Repeat("─", 68), fgBrightBlack))
    fmt.Printf(
        "  %s %s\n\n",
        colored("total wall time", dim),
        colored(fmt.Sprintf("%.2fs", totalWall), bold, fgBrightWhite),
    )
}

type RunResult struct {
    ExitCode    int     `json:"exit_code"`
    Stdout      string  `json:"stdout"`
    Stderr      string  `json:"stderr"`
    WallSecond  float64 `json:"wall_seconds"`
}

type TestResult struct {
    TestName   string    `json:"test_name"`
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
    Name     string       `json:"name"`
    Internal bool         `json:"internal"`
    Category string       `json:"category"`
    Results  []TestResult `json:"results"`
}

type Benchmark struct {
    Name        string  `json:"name"`
    SrcPath     string  `json:"src_path"`
    Category    string  `json:"category"`
    Description string  `json:"description"`
    Internal    bool    `json:"internal"`
}

func runCmd(cwd string, cmd string, args ...string) RunResult {
    start := time.Now()
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
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

func compileGo(src, out string) (float64, string, bool) {
    r := runCmd("", *goCompiler, "build", "-o", out, src)
    return r.WallSecond, r.Stderr, r.ExitCode == 0
}

func compileCpp(src, out string) (float64, string, bool) {
    r := runCmd("", *cppCompiler, "-O2", "-std=c++17", "-pthread", "-o", out, src)
    return r.WallSecond, r.Stderr, r.ExitCode == 0
}

func buildTestSources(testSrc string) ([]string, error) {
	var srcs []string
	srcBase := filepath.Join(SRC, testSrc)
	err := filepath.Walk(
        srcBase,
        func(path string, f os.FileInfo, err error) error {
            if filepath.Ext(path) != "" {
                srcs = append(srcs, path)
            }
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
    }
    return false
}

func mapExt2Lang(path string) string {
    switch filepath.Ext(path) {
        case ".go":
            return "go"
        case ".cpp":
            return "c++"
        case ".py":
            return "python"
        }
    return "unknown"
}

func runLangBenchmark(benchName, lang, src string) TestResult {
    var binary string
    sum := 0.0

    testName := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
    res := TestResult{TestName: testName, Language: lang}

    if skipLang(lang) {
        res.Skipped = true
        res.SkipReason = "filtered lang: " + lang
        goto done
    }

    if _, err := os.Stat(src); err != nil {
        res.Skipped = true
        res.SkipReason = "source not found: " + src
        goto done
    }

    switch lang {
        case "go":
            binary = filepath.Join(BUILD_DIR, benchName + "_go")
            if runtime.GOOS == "windows" {
                binary += ".exe"
            }
            fmt.Printf("%s  %s  %s compiling…",
                clearLine, langTag(lang),
                colored(testName, fgBrightBlack))

            ct, ce, ok := compileGo(src, binary)
            res.CompileS = &ct
            if !ok {
                fmt.Print(clearLine)
                res.Skipped, res.Error, res.SkipReason = true, ce, ce
                return res
            }

        case "c++":
            binary = filepath.Join(BUILD_DIR, benchName + "_cpp")
            if runtime.GOOS == "windows" {
                binary += ".exe"
            }
            fmt.Printf("%s  %s  %s compiling…",
                clearLine, langTag(lang),
                colored(testName, fgBrightBlack))

            ct, ce, ok := compileCpp(src, binary)
            res.CompileS = &ct
            if !ok {
                fmt.Print(clearLine)
                res.Skipped, res.Error, res.SkipReason = true, ce, ce
                goto done
            }

        default:
            v := 0.0
            res.CompileS = &v
    }

    for i := 0; i < *runs; i++ {
        printRunProgress(i, *runs, lang)

        var r RunResult
        if lang == "python" {
            r = runCmd("", "python", src)
        } else {
            r = runCmd("", binary)
        }

        if r.ExitCode != 0 {
            fmt.Print(clearLine)
            res.Skipped, res.Error, res.SkipReason = true, r.Stderr, r.Stderr
            goto done
        }

        res.RunsS = append(res.RunsS, r.WallSecond)
        if i == 0 {
            res.Output = r.Stdout
        }
    }
    fmt.Print(clearLine)

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

done:
    printResultRow(res)
    return res
}

func runBenchmark(b Benchmark) BenchmarkResult {
    res := BenchmarkResult{
        Name:     b.Name,
        Internal: b.Internal,
        Category: b.Category,
        Results:  make([]TestResult, 0),
    }

    srcs, err := buildTestSources(b.SrcPath)
    if err != nil {
        fmt.Printf("  %s reading %s: %v\n", colored("ERROR", bold, fgBrightRed), b.Name, err)
        return res
    }

    for _, src := range srcs {
        lang := mapExt2Lang(src)
        r := runLangBenchmark(b.Name, lang, src)
        res.Results = append(res.Results, r)
    }

    return res
}

func init() {
    flag.Parse()
    colorEnabled = !*noColor

    _ = os.MkdirAll(BUILD_DIR, 0755)
    _ = os.MkdirAll(RESULTS, 0755)
}

func loadBenchmarks() ([]Benchmark, error) {
    var benchmarks []Benchmark
    data, err := os.ReadFile("benchmarks.json")
    if err != nil {
        goto errLoading
    }
    err = json.Unmarshal(data, &benchmarks)
    if err != nil {
        goto errLoading
    }
    return benchmarks, nil
errLoading:
    return nil, err
}

func main() {
    benchmarks, err := loadBenchmarks()
    if err != nil {
        fmt.Printf("%s loading benchmarks: %v\n", colored("ERROR", bold, fgBrightRed), err)
        return
    }

    var active []Benchmark
    for _, b := range benchmarks {
        if *filter == "" || strings.Contains(strings.ToLower(b.Name), strings.ToLower(*filter)) {
            active = append(active, b)
        }
    }

    total := len(active)

    fmt.Printf("\n%s\n", colored("  ▶  benchmark suite", bold, fgBrightWhite))
    fmt.Printf("  %s %s   %s %s   %s %s\n\n",
        colored("runs", dim), colored(fmt.Sprintf("%d", *runs), fgBrightCyan),
        colored("timeout", dim), colored(fmt.Sprintf("%ds", *timeout), fgBrightCyan),
        colored("benchmarks", dim), colored(fmt.Sprintf("%d", total), fgBrightCyan),
    )

    start := time.Now()
    var results []BenchmarkResult

    for i, b := range active {
        printProgress(i, total, b.Name)
        fmt.Println()

        printBenchHeader(i, total, b)

        r := runBenchmark(b)
        results = append(results, r)
    }

    printProgress(total, total, "done")
    fmt.Println()

    wallTime := time.Since(start).Seconds()

    printSummary(results, wallTime)

    payload := map[string]any{
        "timestamp":           time.Now().Unix(),
        "runs_per_benchmark":  *runs,
        "timeout_s":           *timeout,
        "total_wall_s":        wallTime,
        "benchmarks":          results,
    }

    data, _ := json.MarshalIndent(payload, "", "  ")

    outPath := *output
    if outPath == "" {
        outPath = filepath.Join(RESULTS, fmt.Sprintf("run_%d.json", time.Now().Unix()))
    }

    _ = os.WriteFile(outPath, data, 0644)

    fmt.Printf("  %s %s\n\n",
        colored("results →", dim),
        colored(outPath, fgBrightCyan),
    )
}
