/*
Package loader provides functionality to build context and prompts for the LLM.
This file constructs the core system prompt given to the LLM to instruct it on
generating valid, high-coverage Go tests with specific stylistic rules and constraints.

Functions:
- BuildSystemPrompt: Generates the full system prompt using the file context and constraints.
- BuildFewShotExample: Returns several high-quality examples of table-driven Go tests to guide the LLM.
*/
package loader

import (
	"fmt"
	"path/filepath"
	"strings"
)

func BuildSystemPrompt(ctx FileContext) string {
	funcList := "none detected"
	if len(ctx.FuncNames) > 0 {
		funcList = strings.Join(ctx.FuncNames, ", ")
	}

	return fmt.Sprintf(
		"You are an expert Go test engineer. Your ONLY job is to write a complete, compilable Go test file.\n\n"+
			"Rules:\n"+
			"1. Output ONLY a single ```go ... ``` code block. No explanations, no prose.\n"+
			"2. The test file must start with: package %s\n"+
			"3. CRITICAL: Import ONLY packages that YOUR TEST CODE uses, not packages from the source file.\n"+
			"   Your test should typically only need: \"testing\" + any stdlib packages you call (fmt, time, etc).\n"+
			"4. Do NOT import any external packages (packages not in the Go standard library).\n"+
			"5. Do NOT import \"errors\" or \"strings\" unless your TEST explicitly calls errors.New() or strings functions.\n"+
			"6. Every package used must be imported. Check every function call in the TEST code.\n"+
			"7. Every exported function must have at least one test: %s\n"+
			"8. Use table-driven tests. Each row is one branch or edge case.\n"+
			"9. Every test function must start with TestXxx and accept *testing.T.\n\n"+
			"Coverage Requirements (TARGET: 80%%+):\n"+
			"- For every if/else or switch, write one test case per branch.\n"+
			"- For every function returning error, test both success and error paths.\n"+
			"- For numeric input: test zero, negative, and positive cases.\n"+
			"- NEVER hardcode expected values for large inputs (n > 20 for sequences).\n"+
			"  Test PROPERTIES instead: assert fib(n) == fib(n-1) + fib(n-2) in a loop.\n"+
			"- If source has two implementations of the same logic, write a consistency test:\n"+
			"  for i := 0; i < 100; i++ { assert FuncA(i) == FuncB(i) }\n\n"+
			"HTTP & External Dependency Rules:\n"+
			"- NEVER make real network calls. Use net/http/httptest.NewServer() always.\n"+
			"- Only test error cases that the source function ACTUALLY triggers.\n"+
			"  Read the source error conditions carefully before writing expectError:true cases.\n"+
			"- Do not invent error scenarios that cannot be triggered via httptest.\n\n"+
			"--- SOURCE FILE: %s ---\n%s\n--- END SOURCE FILE ---\n\n"+
			"--- EXAMPLES OF PERFECT TEST FILES ---\n%s\n--- END EXAMPLES ---\n\n"+
			"Write the _test.go file now.",
		ctx.PackageName,
		funcList,
		filepath.Base(ctx.FilePath),
		ctx.RawSource,
		BuildFewShotExample(),
	)
}

func BuildFewShotExample() string {
	return "// Example 1: Pure function with table-driven tests\n" +
		"```go\n" +
		"package mathutil\n\n" +
		"import \"testing\"\n\n" +
		"func TestAdd(t *testing.T) {\n" +
		"\tcases := []struct {\n" +
		"\t\tname string\n" +
		"\t\ta, b int\n" +
		"\t\twant int\n" +
		"\t}{\n" +
		"\t\t{\"positive\", 2, 3, 5},\n" +
		"\t\t{\"zero\", 0, 0, 0},\n" +
		"\t\t{\"negative\", -1, -2, -3},\n" +
		"\t}\n" +
		"\tfor _, tc := range cases {\n" +
		"\t\tt.Run(tc.name, func(t *testing.T) {\n" +
		"\t\t\tgot := Add(tc.a, tc.b)\n" +
		"\t\t\tif got != tc.want {\n" +
		"\t\t\t\tt.Errorf(\"Add(%d, %d) = %d; want %d\", tc.a, tc.b, got, tc.want)\n" +
		"\t\t\t}\n" +
		"\t\t})\n" +
		"\t}\n" +
		"}\n" +
		"```\n\n" +
		"// Example 2: Property-based test for mathematical functions (NO hardcoded large values)\n" +
		"```go\n" +
		"package dp\n\n" +
		"import \"testing\"\n\n" +
		"func TestFibProperty(t *testing.T) {\n" +
		"\t// For complex math: test the property, not hardcoded values\n" +
		"\t// This avoids guessing expected values for large n\n" +
		"\tfor n := int64(3); n <= 50; n++ {\n" +
		"\t\tgot := Fib(n)\n" +
		"\t\twant := (Fib(n-1) + Fib(n-2)) % MOD\n" +
		"\t\tif got != want {\n" +
		"\t\t\tt.Errorf(\"Fib(%d) broke recurrence: got %d want %d\", n, got, want)\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n" +
		"```\n\n" +
		"// Example 3: HTTP handler test with httptest (NO real network calls)\n" +
		"```go\n" +
		"package httpclient\n\n" +
		"import (\n" +
		"\t\"fmt\"\n" +
		"\t\"net/http\"\n" +
		"\t\"net/http/httptest\"\n" +
		"\t\"testing\"\n" +
		")\n\n" +
		"func TestGetBody(t *testing.T) {\n" +
		"\tcases := []struct {\n" +
		"\t\tname    string\n" +
		"\t\tcode    int\n" +
		"\t\tbody    string\n" +
		"\t\twantErr bool\n" +
		"\t}{\n" +
		"\t\t{\"success 200\", http.StatusOK, \"hello\", false},\n" +
		"\t\t{\"not found 404\", http.StatusNotFound, \"\", true},\n" +
		"\t}\n" +
		"\tfor _, tc := range cases {\n" +
		"\t\tt.Run(tc.name, func(t *testing.T) {\n" +
		"\t\t\tts := httptest.NewServer(http.HandlerFunc(\n" +
		"\t\t\t\tfunc(w http.ResponseWriter, r *http.Request) {\n" +
		"\t\t\t\t\tw.WriteHeader(tc.code)\n" +
		"\t\t\t\t\tfmt.Fprint(w, tc.body)\n" +
		"\t\t\t\t}))\n" +
		"\t\t\tdefer ts.Close()\n" +
		"\t\t\t_, err := GetBody(ts.URL)\n" +
		"\t\t\tif (err != nil) != tc.wantErr {\n" +
		"\t\t\t\tt.Errorf(\"GetBody() error = %v, wantErr %v\", err, tc.wantErr)\n" +
		"\t\t\t}\n" +
		"\t\t})\n" +
		"\t}\n" +
		"}\n" +
		"```"
}
