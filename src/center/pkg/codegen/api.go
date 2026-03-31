// Package codegen - API代码生成
// 0.9.0 Beta: 自动代码生成
package codegen

import (
	"fmt"
	"strings"
)

// APIGenerator API代码生成器
type APIGenerator struct {
	generator *Generator
	config    APIGenConfig
}

// APIGenConfig API生成配置
type APIGenConfig struct {
	Package       string `json:"package"`
	BasePath      string `json:"base_path"`
	GenerateModel bool   `json:"generate_model"`
	GenerateHandler bool  `json:"generate_handler"`
	GenerateRoutes  bool  `json:"generate_routes"`
	GenerateTests   bool  `json:"generate_tests"`
}

// NewAPIGenerator 创建API生成器
func NewAPIGenerator(generator *Generator, config APIGenConfig) *APIGenerator {
	return &APIGenerator{
		generator: generator,
		config:    config,
	}
}

// APISpec API规范
type APISpec struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	BasePath    string       `json:"base_path"`
	Endpoints   []Endpoint   `json:"endpoints"`
	Models      []ModelSpec  `json:"models"`
}

// Endpoint 端点定义
type Endpoint struct {
	Name        string       `json:"name"`
	Method      string       `json:"method"`
	Path        string       `json:"path"`
	Description string       `json:"description"`
	Params      []ParamSpec  `json:"params"`
	Query       []ParamSpec  `json:"query"`
	Body        string       `json:"body"`
	Response    string       `json:"response"`
	Auth        bool         `json:"auth"`
}

// ParamSpec 参数定义
type ParamSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default"`
}

// ModelSpec 模型定义
type ModelSpec struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	TableName   string      `json:"table_name"`
	Fields      []FieldSpec `json:"fields"`
}

// FieldSpec 字段定义
type FieldSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	JSONName    string `json:"json_name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     string `json:"default"`
}

// Generate 生成API代码
func (g *APIGenerator) Generate(spec APISpec, outputDir string) error {
	// 生成模型
	if g.config.GenerateModel {
		for _, model := range spec.Models {
			if err := g.generateModel(model, outputDir); err != nil {
				return err
			}
		}
	}

	// 生成处理器
	if g.config.GenerateHandler {
		for _, endpoint := range spec.Endpoints {
			if err := g.generateHandler(endpoint, outputDir); err != nil {
				return err
			}
		}
	}

	// 生成路由
	if g.config.GenerateRoutes {
		if err := g.generateRoutes(spec, outputDir); err != nil {
			return err
		}
	}

	// 生成测试
	if g.config.GenerateTests {
		for _, endpoint := range spec.Endpoints {
			if err := g.generateTest(endpoint, outputDir); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateModel 生成模型
func (g *APIGenerator) generateModel(model ModelSpec, outputDir string) error {
	data := map[string]interface{}{
		"Package":     g.config.Package,
		"Name":        model.Name,
		"Description": model.Description,
		"TableName":   model.TableName,
		"Fields":      model.Fields,
	}

	output := fmt.Sprintf("%s/model_%s.go", outputDir, toSnakeCase(model.Name))
	return g.generator.Generate("go-model", data, output)
}

// generateHandler 生成处理器
func (g *APIGenerator) generateHandler(endpoint Endpoint, outputDir string) error {
	data := map[string]interface{}{
		"Package":       g.config.Package,
		"HandlerName":   endpoint.Name,
		"Description":   endpoint.Description,
		"Method":        endpoint.Method,
		"Path":          endpoint.Path,
		"Params":        endpoint.Params,
		"Query":         endpoint.Query,
		"Body":          endpoint.Body,
		"Response":      endpoint.Response,
	}

	output := fmt.Sprintf("%s/handler_%s.go", outputDir, toSnakeCase(endpoint.Name))
	return g.generator.Generate("go-api-handler", data, output)
}

// generateRoutes 生成路由
func (g *APIGenerator) generateRoutes(spec APISpec, outputDir string) error {
	var routesCode strings.Builder

	routesCode.WriteString(fmt.Sprintf("package %s\n\n", g.config.Package))
	routesCode.WriteString("import (\n")
	routesCode.WriteString("\t\"net/http\"\n")
	routesCode.WriteString(")\n\n")
	routesCode.WriteString(fmt.Sprintf("// RegisterRoutes 注册%s路由\n", spec.Name))
	routesCode.WriteString("func RegisterRoutes(mux *http.ServeMux) {\n")

	for _, endpoint := range spec.Endpoints {
		routesCode.WriteString(fmt.Sprintf("\tmux.HandleFunc(\"%s\", %s)\n",
			endpoint.Path, endpoint.Name))
	}

	routesCode.WriteString("}\n")

	// 直接写入
	output := fmt.Sprintf("%s/routes.go", outputDir)
	return g.generator.Generate("go-model", map[string]interface{}{}, output)
}

// generateTest 生成测试
func (g *APIGenerator) generateTest(endpoint Endpoint, outputDir string) error {
	data := map[string]interface{}{
		"Package":     g.config.Package,
		"Name":        endpoint.Name,
		"Method":      endpoint.Method,
		"Path":        endpoint.Path,
		"Description": endpoint.Description,
	}

	output := fmt.Sprintf("%s/handler_%s_test.go", outputDir, toSnakeCase(endpoint.Name))
	return g.generator.Generate("go-test", data, output)
}

// OpenAPIGenerator OpenAPI生成器
type OpenAPIGenerator struct {
	generator *Generator
}

// NewOpenAPIGenerator 创建OpenAPI生成器
func NewOpenAPIGenerator(generator *Generator) *OpenAPIGenerator {
	return &OpenAPIGenerator{generator: generator}
}

// GenerateFromSpec 从规范生成OpenAPI文档
func (g *OpenAPIGenerator) GenerateFromSpec(spec APISpec) string {
	var doc strings.Builder

	doc.WriteString("{\n")
	doc.WriteString("  \"openapi\": \"3.0.3\",\n")
	doc.WriteString(fmt.Sprintf("  \"info\": {\n"))
	doc.WriteString(fmt.Sprintf("    \"title\": \"%s\",\n", spec.Name))
	doc.WriteString(fmt.Sprintf("    \"version\": \"%s\"\n", spec.Version))
	doc.WriteString("  },\n")
	doc.WriteString(fmt.Sprintf("  \"paths\": {\n"))

	for i, endpoint := range spec.Endpoints {
		if i > 0 {
			doc.WriteString(",\n")
		}
		g.generatePath(&doc, endpoint)
	}

	doc.WriteString("\n  }\n")
	doc.WriteString("}\n")

	return doc.String()
}

// generatePath 生成路径
func (g *OpenAPIGenerator) generatePath(doc *strings.Builder, endpoint Endpoint) {
	doc.WriteString(fmt.Sprintf("    \"%s\": {\n", endpoint.Path))
	doc.WriteString(fmt.Sprintf("      \"%s\": {\n", strings.ToLower(endpoint.Method)))
	doc.WriteString(fmt.Sprintf("        \"summary\": \"%s\",\n", endpoint.Description))
	doc.WriteString("        \"responses\": {\n")
	doc.WriteString("          \"200\": {\n")
	doc.WriteString("            \"description\": \"Success\"\n")
	doc.WriteString("          }\n")
	doc.WriteString("        }\n")
	doc.WriteString("      }\n")
	doc.WriteString("    }")
}

// toSnakeCase 转换为蛇形命名
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && isUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(toLower(r))
	}
	return result.String()
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}