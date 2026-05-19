package loader

import (
	"fmt"
	"path/filepath"
	"strings"
)

// BuildSystemPrompt constructs the full LLM system prompt for test generation.
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
			"3. Import \"testing\" and ANY standard library packages your test code uses (e.g., \"fmt\", \"strings\", \"time\", etc.).\n"+
			"4. Do NOT import any external packages (packages not in the Go standard library).\n"+
			"5. CRITICAL: If you use fmt.Errorf, fmt.Sprintf, or any fmt functions, you MUST import \"fmt\" in the imports block.\n"+
			"6. CRITICAL: Check every function and method call in your test — ensure every package it comes from is imported.\n"+
			"7. Every exported function must have at least one test: %s\n"+
			"8. Use table-driven tests where applicable.\n"+
			"9. Every test function must start with TestXxx and accept *testing.T as its only argument.\n\n"+
			"Coverage Requirements (TARGET: 80%%+):\n"+
			"- Every exported function must have at least one test.\n"+
			"- For every if/else or switch, write one test case per branch.\n"+
			"- For every function returning error, test both success and error paths.\n"+
			"- For numeric input: always test zero, negative, and positive cases.\n"+
			"- Each row in a table-driven test is a branch or edge case.\n\n"+
			"--- SOURCE FILE: %s ---\n%s\n--- END SOURCE FILE ---\n\n"+
			"--- EXAMPLE OF A PERFECT TEST FILE ---\n%s\n--- END EXAMPLE ---\n\n"+
			"Write the _test.go file now.",
		ctx.PackageName,
		funcList,
		filepath.Base(ctx.FilePath),
		ctx.RawSource,
		BuildFewShotExample(),
	)
}

// BuildFewShotExample returns a hardcoded gold-standard test file example.
func BuildFewShotExample() string {
	return "```go\n" +
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
		"```"
}
