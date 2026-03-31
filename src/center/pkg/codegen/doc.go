// Package codegen - 文档生成
// Sprint 31: v9.0 自动代码生成
package codegen

import (
	"fmt"
	"strings"
)

// DocGenerator 文档生成器
type DocGenerator struct {
	generator *Generator
}

// NewDocGenerator 创建文档生成器
func NewDocGenerator(generator *Generator) *DocGenerator {
	return &DocGenerator{generator: generator}
}

// DocSpec 文档规范
type DocSpec struct {
	Title       string        `json:"title"`
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Sections    []DocSection  `json:"sections"`
	APIs        []APIDoc      `json:"apis"`
}

// DocSection 文档章节
type DocSection struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	SubSections []DocSection `json:"subsections"`
}

// APIDoc API文档
type APIDoc struct {
	Name        string       `json:"name"`
	Method      string       `json:"method"`
	Path        string       `json:"path"`
	Description string       `json:"description"`
	Params      []ParamDoc   `json:"params"`
	Response    ResponseDoc  `json:"response"`
	Examples    []ExampleDoc `json:"examples"`
}

// ParamDoc 参数文档
type ParamDoc struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	In          string `json:"in"` // path, query, body, header
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// ResponseDoc 响应文档
type ResponseDoc struct {
	Code        int               `json:"code"`
	Description string            `json:"description"`
	Fields      []ResponseField   `json:"fields"`
}

// ResponseField 响应字段
type ResponseField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ExampleDoc 示例文档
type ExampleDoc struct {
	Title   string `json:"title"`
	Request string `json:"request"`
	Response string `json:"response"`
}

// GenerateMarkdown 生成Markdown文档
func (g *DocGenerator) GenerateMarkdown(spec DocSpec) string {
	var md strings.Builder

	// 标题
	md.WriteString(fmt.Sprintf("# %s\n\n", spec.Title))
	md.WriteString(fmt.Sprintf("Version: %s\n\n", spec.Version))
	md.WriteString(fmt.Sprintf("%s\n\n", spec.Description))

	// 目录
	md.WriteString("## 目录\n\n")
	for i, section := range spec.Sections {
		md.WriteString(fmt.Sprintf("%d. [%s](#%s)\n", i+1, section.Title, toAnchor(section.Title)))
	}
	md.WriteString("\n")

	// 章节
	for _, section := range spec.Sections {
		g.generateSection(&md, section, 2)
	}

	// API文档
	if len(spec.APIs) > 0 {
		md.WriteString("## API参考\n\n")
		for _, api := range spec.APIs {
			g.generateAPIDoc(&md, api)
		}
	}

	return md.String()
}

// generateSection 生成章节
func (g *DocGenerator) generateSection(md *strings.Builder, section DocSection, level int) {
	md.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), section.Title))
	md.WriteString(fmt.Sprintf("%s\n\n", section.Content))

	for _, sub := range section.SubSections {
		g.generateSection(md, sub, level+1)
	}
}

// generateAPIDoc 生成API文档
func (g *DocGenerator) generateAPIDoc(md *strings.Builder, api APIDoc) {
	md.WriteString(fmt.Sprintf("### %s\n\n", api.Name))
	md.WriteString(fmt.Sprintf("`%s %s`\n\n", api.Method, api.Path))
	md.WriteString(fmt.Sprintf("%s\n\n", api.Description))

	// 参数
	if len(api.Params) > 0 {
		md.WriteString("**参数**\n\n")
		md.WriteString("| 名称 | 类型 | 位置 | 必需 | 描述 |\n")
		md.WriteString("|------|------|------|------|------|\n")
		for _, param := range api.Params {
			md.WriteString(fmt.Sprintf("| %s | %s | %s | %v | %s |\n",
				param.Name, param.Type, param.In, param.Required, param.Description))
		}
		md.WriteString("\n")
	}

	// 响应
	md.WriteString("**响应**\n\n")
	md.WriteString(fmt.Sprintf("状态码: %d\n\n", api.Response.Code))
	md.WriteString(fmt.Sprintf("%s\n\n", api.Response.Description))

	if len(api.Response.Fields) > 0 {
		md.WriteString("| 字段 | 类型 | 描述 |\n")
		md.WriteString("|------|------|------|\n")
		for _, field := range api.Response.Fields {
			md.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
				field.Name, field.Type, field.Description))
		}
		md.WriteString("\n")
	}

	// 示例
	if len(api.Examples) > 0 {
		md.WriteString("**示例**\n\n")
		for _, example := range api.Examples {
			md.WriteString(fmt.Sprintf("%s\n\n", example.Title))
			md.WriteString("请求:\n")
			md.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", example.Request))
			md.WriteString("响应:\n")
			md.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", example.Response))
		}
	}
}

// GenerateHTML 生成HTML文档
func (g *DocGenerator) GenerateHTML(spec DocSpec) string {
	var html strings.Builder

	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html lang=\"zh-CN\">\n")
	html.WriteString("<head>\n")
	html.WriteString(fmt.Sprintf("  <title>%s</title>\n", spec.Title))
	html.WriteString("  <meta charset=\"utf-8\">\n")
	html.WriteString("  <style>\n")
	html.WriteString("    body { font-family: Arial, sans-serif; margin: 40px; }\n")
	html.WriteString("    h1 { color: #333; }\n")
	html.WriteString("    h2 { color: #666; margin-top: 30px; }\n")
	html.WriteString("    code { background: #f4f4f4; padding: 2px 6px; }\n")
	html.WriteString("    pre { background: #f4f4f4; padding: 15px; overflow-x: auto; }\n")
	html.WriteString("    table { border-collapse: collapse; width: 100%; margin: 20px 0; }\n")
	html.WriteString("    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }\n")
	html.WriteString("    th { background: #f4f4f4; }\n")
	html.WriteString("  </style>\n")
	html.WriteString("</head>\n")
	html.WriteString("<body>\n")

	html.WriteString(fmt.Sprintf("<h1>%s</h1>\n", spec.Title))
	html.WriteString(fmt.Sprintf("<p>Version: %s</p>\n", spec.Version))
	html.WriteString(fmt.Sprintf("<p>%s</p>\n", spec.Description))

	// 章节
	for _, section := range spec.Sections {
		g.generateHTMLSection(&html, section, 2)
	}

	// API文档
	if len(spec.APIs) > 0 {
		html.WriteString("<h2>API参考</h2>\n")
		for _, api := range spec.APIs {
			g.generateHTMLAPI(&html, api)
		}
	}

	html.WriteString("</body>\n")
	html.WriteString("</html>\n")

	return html.String()
}

// generateHTMLSection 生成HTML章节
func (g *DocGenerator) generateHTMLSection(html *strings.Builder, section DocSection, level int) {
	html.WriteString(fmt.Sprintf("<h%d>%s</h%d>\n", level, section.Title, level))
	html.WriteString(fmt.Sprintf("<p>%s</p>\n", section.Content))

	for _, sub := range section.SubSections {
		g.generateHTMLSection(html, sub, level+1)
	}
}

// generateHTMLAPI 生成HTML API文档
func (g *DocGenerator) generateHTMLAPI(html *strings.Builder, api APIDoc) {
	html.WriteString(fmt.Sprintf("<h3>%s</h3>\n", api.Name))
	html.WriteString(fmt.Sprintf("<code>%s %s</code>\n", api.Method, api.Path))
	html.WriteString(fmt.Sprintf("<p>%s</p>\n", api.Description))

	// 参数表格
	if len(api.Params) > 0 {
		html.WriteString("<h4>参数</h4>\n")
		html.WriteString("<table>\n")
		html.WriteString("<tr><th>名称</th><th>类型</th><th>位置</th><th>必需</th><th>描述</th></tr>\n")
		for _, param := range api.Params {
			html.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%v</td><td>%s</td></tr>\n",
				param.Name, param.Type, param.In, param.Required, param.Description))
		}
		html.WriteString("</table>\n")
	}
}

// GenerateOpenAPI 生成OpenAPI文档
func (g *DocGenerator) GenerateOpenAPI(spec DocSpec) string {
	var openapi strings.Builder

	openapi.WriteString("{\n")
	openapi.WriteString("  \"openapi\": \"3.0.3\",\n")
	openapi.WriteString(fmt.Sprintf("  \"info\": {\n"))
	openapi.WriteString(fmt.Sprintf("    \"title\": \"%s\",\n", spec.Title))
	openapi.WriteString(fmt.Sprintf("    \"version\": \"%s\",\n", spec.Version))
	openapi.WriteString(fmt.Sprintf("    \"description\": \"%s\"\n", spec.Description))
	openapi.WriteString("  },\n")
	openapi.WriteString("  \"paths\": {\n")

	for i, api := range spec.APIs {
		if i > 0 {
			openapi.WriteString(",\n")
		}
		openapi.WriteString(fmt.Sprintf("    \"%s\": {\n", api.Path))
		openapi.WriteString(fmt.Sprintf("      \"%s\": {\n", strings.ToLower(api.Method)))
		openapi.WriteString(fmt.Sprintf("        \"summary\": \"%s\",\n", api.Name))
		openapi.WriteString(fmt.Sprintf("        \"description\": \"%s\",\n", api.Description))

		// 参数
		if len(api.Params) > 0 {
			openapi.WriteString("        \"parameters\": [\n")
			for j, param := range api.Params {
				if j > 0 {
					openapi.WriteString(",\n")
				}
				openapi.WriteString("          {\n")
				openapi.WriteString(fmt.Sprintf("            \"name\": \"%s\",\n", param.Name))
				openapi.WriteString(fmt.Sprintf("            \"in\": \"%s\",\n", param.In))
				openapi.WriteString(fmt.Sprintf("            \"required\": %v,\n", param.Required))
				openapi.WriteString(fmt.Sprintf("            \"schema\": { \"type\": \"%s\" }\n", param.Type))
				openapi.WriteString("          }")
			}
			openapi.WriteString("\n        ],\n")
		}

		// 响应
		openapi.WriteString("        \"responses\": {\n")
		openapi.WriteString(fmt.Sprintf("          \"%d\": {\n", api.Response.Code))
		openapi.WriteString(fmt.Sprintf("            \"description\": \"%s\"\n", api.Response.Description))
		openapi.WriteString("          }\n")
		openapi.WriteString("        }\n")

		openapi.WriteString("      }\n")
		openapi.WriteString("    }")
	}

	openapi.WriteString("\n  }\n")
	openapi.WriteString("}\n")

	return openapi.String()
}

// GenerateReadme 生成README文档
func (g *DocGenerator) GenerateReadme(spec DocSpec) string {
	var readme strings.Builder

	readme.WriteString(fmt.Sprintf("# %s\n\n", spec.Title))
	readme.WriteString(fmt.Sprintf("%s\n\n", spec.Description))
	readme.WriteString(fmt.Sprintf("![Version](https://img.shields.io/badge/version-%s-blue)\n\n", spec.Version))

	readme.WriteString("## 快速开始\n\n")
	readme.WriteString("### 安装\n\n")
	readme.WriteString("```bash\n")
	readme.WriteString("go get github.com/ofa/sdk\n")
	readme.WriteString("```\n\n")

	readme.WriteString("### 使用\n\n")
	readme.WriteString("```go\n")
	readme.WriteString("package main\n\n")
	readme.WriteString("import \"github.com/ofa/sdk\"\n\n")
	readme.WriteString("func main() {\n")
	readme.WriteString("    client := sdk.NewClient(\"http://localhost:8080\", \"your-api-key\")\n")
	readme.WriteString("    // 使用客户端...\n")
	readme.WriteString("}\n")
	readme.WriteString("```\n\n")

	readme.WriteString("## 文档\n\n")
	for _, section := range spec.Sections {
		readme.WriteString(fmt.Sprintf("- [%s](#%s)\n", section.Title, toAnchor(section.Title)))
	}
	readme.WriteString("\n")

	readme.WriteString("## 许可证\n\n")
	readme.WriteString("Apache 2.0\n")

	return readme.String()
}

// toAnchor 转换为锚点
func toAnchor(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	return s
}