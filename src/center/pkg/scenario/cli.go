// Package scenario - 场景验证命令行工具
package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CLIConfig 命令行配置
type CLIConfig struct {
	OutputPath string
	Format     string // json, text, html
	Scenarios  []string
	Verbose    bool
}

// RunValidationCLI 运行验证命令行
func RunValidationCLI(cliConfig CLIConfig) error {
	config := ScenarioConfig{
		MockMode:     true,
		Timeout:      5 * time.Minute,
		TestDuration: 30 * time.Second,
		Verbose:      cliConfig.Verbose,
		ReportPath:   cliConfig.OutputPath,
	}

	validator := NewScenarioValidator(config)
	ctx := context.Background()

	fmt.Println("=== OFA 场景验证测试 ===")
	fmt.Printf("开始时间: %s\n", time.Now().Format(time.RFC3339))

	var report *ValidationReport

	if len(cliConfig.Scenarios) > 0 {
		// 运行指定场景
		report = &ValidationReport{
			StartTime: time.Now(),
			Scenarios: make(map[string]*ScenarioResult),
			Summary:   &ValidationSummary{},
		}

		for _, name := range cliConfig.Scenarios {
			fmt.Printf("\n运行场景: %s\n", name)
			result, err := validator.RunScenario(ctx, name)
			if err != nil {
				fmt.Printf("  错误: %v\n", err)
				continue
			}
			report.Scenarios[name] = result
			report.Summary.TotalTests += result.TestCount
			report.Summary.TotalPassed += result.PassedCount
			report.Summary.TotalFailed += result.FailedCount
			if result.Status == "passed" {
				report.Summary.ScenariosPassed++
			}
		}
		report.Summary.ScenariosTotal = len(report.Scenarios)
		report.Summary.PassRate = float64(report.Summary.TotalPassed) / float64(report.Summary.TotalTests) * 100
	} else {
		// 运行所有场景
		fmt.Println("\n运行所有场景...")
		report = validator.RunAllScenarios(ctx)
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	// 输出报告
	switch cliConfig.Format {
	case "json":
		outputJSON(report, cliConfig.OutputPath)
	case "html":
		outputHTML(report, cliConfig.OutputPath)
	default:
		outputText(report, cliConfig.Verbose)
	}

	// 保存报告
	if cliConfig.OutputPath != "" {
		saveReport(report, cliConfig.OutputPath)
	}

	return nil
}

// outputJSON 输出JSON格式
func outputJSON(report *ValidationReport, outputPath string) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("JSON序列化错误: %v\n", err)
		return
	}

	if outputPath != "" {
		fmt.Printf("报告已保存到: %s\n", outputPath)
	} else {
		fmt.Println(string(data))
	}
}

// outputText 输出文本格式
func outputText(report *ValidationReport, verbose bool) {
	fmt.Println(report.PrintReport())

	if verbose {
		for name, result := range report.Scenarios {
			fmt.Printf("\n场景详情: %s\n", name)
			for _, detail := range result.Details {
				status := "✅"
				if detail.Status == "failed" {
					status = "❌"
				}
				fmt.Printf("  %s %s: %s (%v)\n", status, detail.TestName, detail.Description, detail.Duration)
				if detail.Error != "" {
					fmt.Printf("    错误: %s\n", detail.Error)
				}
				if detail.Data != nil {
					data, _ := json.MarshalIndent(detail.Data, "    ", "  ")
					fmt.Printf("    数据: %s\n", string(data))
				}
			}
		}
	}
}

// outputHTML 输出HTML格式
func outputHTML(report *ValidationReport, outputPath string) {
	html := generateHTMLReport(report)
	if outputPath != "" {
		os.WriteFile(outputPath, []byte(html), 0644)
		fmt.Printf("HTML报告已保存到: %s\n", outputPath)
	} else {
		fmt.Println(html)
	}
}

// generateHTMLReport 生成HTML报告
func generateHTMLReport(report *ValidationReport) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>OFA 场景验证报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        .summary { background: #f5f5f5; padding: 15px; border-radius: 5px; }
        .scenario { margin: 20px 0; padding: 15px; border: 1px solid #ddd; }
        .passed { background: #e8f5e9; }
        .failed { background: #ffebee; }
        .test { margin: 10px 0; padding: 10px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>OFA 场景验证报告</h1>
    <div class="summary">
        <h2>汇总</h2>
        <p>测试时间: %s</p>
        <p>总耗时: %v</p>
        <p>场景通过: %d/%d</p>
        <p>测试通过: %d/%d</p>
        <p>通过率: %.2f%%</p>
    </div>
    <div class="scenarios">
        <h2>场景详情</h2>
        %s
    </div>
</body>
</html>`,
		report.StartTime.Format(time.RFC3339),
		report.Duration,
		report.Summary.ScenariosPassed, report.Summary.ScenariosTotal,
		report.Summary.TotalPassed, report.Summary.TotalTests,
		report.Summary.PassRate,
		generateScenarioHTML(report),
	)
}

// generateScenarioHTML 生成场景HTML
func generateScenarioHTML(report *ValidationReport) string {
	var html string
	for name, result := range report.Scenarios {
		class := "passed"
		if result.Status == "failed" {
			class = "failed"
		}
		html += fmt.Sprintf(`<div class="scenario %s">
    <h3>%s</h3>
    <p>状态: %s | 通过: %d/%d | 耗时: %v</p>
    <div class="tests">
`, class, name, result.Status, result.PassedCount, result.TestCount, result.Duration)

		for _, detail := range result.Details {
			html += fmt.Sprintf(`        <div class="test %s">
            <strong>%s</strong>: %s (%v)
        </div>
`, detail.Status, detail.TestName, detail.Description, detail.Duration)
		}
		html += "    </div>\n</div>\n"
	}
	return html
}

// saveReport 保存报告
func saveReport(report *ValidationReport, outputPath string) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("保存报告错误: %v\n", err)
		return
	}

	// 确保目录存在
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0755)
	}

	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		fmt.Printf("写入文件错误: %v\n", err)
		return
	}

	fmt.Printf("报告已保存: %s\n", outputPath)
}

// QuickValidation 快速验证（用于快速检查）
func QuickValidation() bool {
	config := ScenarioConfig{
		MockMode:     true,
		Timeout:      30 * time.Second,
		TestDuration: 5 * time.Second,
	}

	validator := NewScenarioValidator(config)
	ctx := context.Background()

	report := validator.RunAllScenarios(ctx)

	fmt.Println(report.PrintReport())
	return report.Summary.PassRate == 100.0
}

// Main 主函数入口
func Main() {
	cliConfig := CLIConfig{
		Format:  "text",
		Verbose: true,
	}

	// 解析命令行参数
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				cliConfig.OutputPath = args[i+1]
				i++
			}
		case "-f", "--format":
			if i+1 < len(args) {
				cliConfig.Format = args[i+1]
				i++
			}
		case "-s", "--scenario":
			if i+1 < len(args) {
				cliConfig.Scenarios = append(cliConfig.Scenarios, args[i+1])
				i++
			}
		case "-v", "--verbose":
			cliConfig.Verbose = true
		case "-q", "--quiet":
			cliConfig.Verbose = false
		case "-h", "--help":
			printHelp()
			return
		}
	}

	RunValidationCLI(cliConfig)
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println(`OFA 场景验证工具

用法: validator [选项]

选项:
  -o, --output <path>   输出报告路径
  -f, --format <format> 输出格式 (json/text/html)
  -s, --scenario <name> 运行指定场景 (可多次使用)
  -v, --verbose         详细输出
  -q, --quiet           简洁输出
  -h, --help            显示帮助

场景名称:
  cross_device        跨设备协同测试
  smart_home          智能家居联动测试
  distributed_ai      分布式AI推理测试
  privacy_computing   隐私计算验证

示例:
  validator                    # 运行所有场景
  validator -s cross_device    # 仅运行跨设备场景
  validator -o report.json -f json  # 保存JSON报告`)
}