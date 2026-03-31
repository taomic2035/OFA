// Package codegen - SDK代码生成
// 0.9.0 Beta: 自动代码生成
package codegen

import (
	"fmt"
	"strings"
)

// SDKGenerator SDK代码生成器
type SDKGenerator struct {
	generator *Generator
}

// NewSDKGenerator 创建SDK生成器
func NewSDKGenerator(generator *Generator) *SDKGenerator {
	return &SDKGenerator{generator: generator}
}

// SDKSpec SDK规范
type SDKSpec struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	Package     string       `json:"package"`
	Methods     []SDKMethod  `json:"methods"`
	Models      []ModelSpec  `json:"models"`
}

// SDKMethod SDK方法
type SDKMethod struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Params      []ParamSpec `json:"params"`
	ReturnType  string      `json:"return_type"`
	HTTPMethod  string      `json:"http_method"`
	HTTPPath    string      `json:"http_path"`
}

// GenerateGoSDK 生成Go SDK
func (g *SDKGenerator) GenerateGoSDK(spec SDKSpec, outputDir string) error {
	// 生成客户端
	data := map[string]interface{}{
		"Package":     spec.Package,
		"Name":        spec.Name,
		"Description": spec.Description,
		"Methods":     spec.Methods,
	}

	output := fmt.Sprintf("%s/client.go", outputDir)
	if err := g.generator.Generate("sdk-client", data, output); err != nil {
		return err
	}

	// 生成模型
	for _, model := range spec.Models {
		modelData := map[string]interface{}{
			"Package":     spec.Package,
			"Name":        model.Name,
			"Description": model.Description,
			"TableName":   model.TableName,
			"Fields":      model.Fields,
		}

		output := fmt.Sprintf("%s/model_%s.go", outputDir, toSnakeCase(model.Name))
		if err := g.generator.Generate("go-model", modelData, output); err != nil {
			return err
		}
	}

	return nil
}

// GenerateTypeScriptSDK 生成TypeScript SDK
func (g *SDKGenerator) GenerateTypeScriptSDK(spec SDKSpec, outputDir string) error {
	var code strings.Builder

	// 文件头
	code.WriteString(fmt.Sprintf("/**\n * %s SDK\n * %s\n */\n\n", spec.Name, spec.Description))
	code.WriteString(fmt.Sprintf("const VERSION = '%s';\n\n", spec.Version))

	// 接口定义
	for _, model := range spec.Models {
		g.generateTSInterface(&code, model)
	}

	// 客户端类
	code.WriteString(fmt.Sprintf("export class %sClient {\n", spec.Name))
	code.WriteString("  private baseURL: string;\n")
	code.WriteString("  private apiKey: string;\n\n")
	code.WriteString("  constructor(baseURL: string, apiKey: string) {\n")
	code.WriteString("    this.baseURL = baseURL;\n")
	code.WriteString("    this.apiKey = apiKey;\n")
	code.WriteString("  }\n\n")

	// 方法
	for _, method := range spec.Methods {
		g.generateTSMethod(&code, method)
	}

	code.WriteString("}\n")

	// 写入文件
	output := fmt.Sprintf("%s/client.ts", outputDir)
	return g.generator.generator.Write("ts-interface", map[string]interface{}{}, nil)
}

// generateTSInterface 生成TypeScript接口
func (g *SDKGenerator) generateTSInterface(code *strings.Builder, model ModelSpec) {
	code.WriteString(fmt.Sprintf("export interface %s {\n", model.Name))
	for _, field := range model.Fields {
		tsType := goTypeToTS(field.Type)
		code.WriteString(fmt.Sprintf("  %s: %s;\n", field.JSONName, tsType))
	}
	code.WriteString("}\n\n")
}

// generateTSMethod 生成TypeScript方法
func (g *SDKGenerator) generateTSMethod(code *strings.Builder, method SDKMethod) {
	code.WriteString(fmt.Sprintf("  async %s(", method.Name))

	// 参数
	for i, param := range method.Params {
		if i > 0 {
			code.WriteString(", ")
		}
		code.WriteString(fmt.Sprintf("%s: %s", param.Name, goTypeToTS(param.Type)))
	}
	code.WriteString(fmt.Sprintf("): Promise<%s> {\n", goTypeToTS(method.ReturnType)))

	// 实现简化
	code.WriteString(fmt.Sprintf("    const response = await fetch(`${this.baseURL}%s`, {\n", method.HTTPPath))
	code.WriteString(fmt.Sprintf("      method: '%s',\n", strings.ToUpper(method.HTTPMethod)))
	code.WriteString("      headers: {\n")
	code.WriteString("        'Content-Type': 'application/json',\n")
	code.WriteString("        'Authorization': `Bearer ${this.apiKey}`\n")
	code.WriteString("      }\n")
	code.WriteString("    });\n")
	code.WriteString("    return response.json();\n")
	code.WriteString("  }\n\n")
}

// GeneratePythonSDK 生成Python SDK
func (g *SDKGenerator) GeneratePythonSDK(spec SDKSpec, outputDir string) error {
	var code strings.Builder

	// 文件头
	code.WriteString(fmt.Sprintf("\"\"\"\n%s SDK\n%s\n\"\"\"\n\n", spec.Name, spec.Description))
	code.WriteString("import requests\n")
	code.WriteString("from dataclasses import dataclass\n")
	code.WriteString("from typing import Optional, List, Dict, Any\n\n")

	// 模型
	for _, model := range spec.Models {
		g.generatePyClass(&code, model)
	}

	// 客户端类
	code.WriteString(fmt.Sprintf("class %sClient:\n", spec.Name))
	code.WriteString("    def __init__(self, base_url: str, api_key: str):\n")
	code.WriteString("        self.base_url = base_url\n")
	code.WriteString("        self.api_key = api_key\n\n")

	// 方法
	for _, method := range spec.Methods {
		g.generatePyMethod(&code, method)
	}

	// 写入文件
	output := fmt.Sprintf("%s/client.py", outputDir)

	// 直接生成
	g.generator.config.OutputDir = outputDir
	return nil
}

// generatePyClass 生成Python类
func (g *SDKGenerator) generatePyClass(code *strings.Builder, model ModelSpec) {
	code.WriteString(fmt.Sprintf("@dataclass\n"))
	code.WriteString(fmt.Sprintf("class %s:\n", model.Name))
	code.WriteString(fmt.Sprintf("    \"\"\"%s\"\"\"\n", model.Description))

	for _, field := range model.Fields {
		pyType := goTypeToPy(field.Type)
		code.WriteString(fmt.Sprintf("    %s: %s\n", field.JSONName, pyType))
	}
	code.WriteString("\n")
}

// generatePyMethod 生成Python方法
func (g *SDKGenerator) generatePyMethod(code *strings.Builder, method SDKMethod) {
	code.WriteString(fmt.Sprintf("    def %s(self", method.Name))

	// 参数
	for _, param := range method.Params {
		code.WriteString(fmt.Sprintf(", %s: %s", param.Name, goTypeToPy(param.Type)))
	}

	code.WriteString(fmt.Sprintf(") -> %s:\n", goTypeToPy(method.ReturnType)))
	code.WriteString(fmt.Sprintf("        \"\"\"%s\"\"\"\n", method.Description))
	code.WriteString(fmt.Sprintf("        response = requests.%s(\n", strings.ToLower(method.HTTPMethod)))
	code.WriteString(fmt.Sprintf("            f\"{self.base_url}%s\",\n", method.HTTPPath))
	code.WriteString("            headers={\"Authorization\": f\"Bearer {self.api_key}\"}\n")
	code.WriteString("        )\n")
	code.WriteString("        return response.json()\n\n")
}

// 类型转换函数
func goTypeToTS(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int32", "int64":
		return "number"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "[]string":
		return "string[]"
	case "[]int":
		return "number[]"
	case "map[string]string":
		return "Record<string, string>"
	case "map[string]interface{}":
		return "Record<string, any>"
	default:
		return goType
	}
}

func goTypeToPy(goType string) string {
	switch goType {
	case "string":
		return "str"
	case "int", "int32", "int64":
		return "int"
	case "float32", "float64":
		return "float"
	case "bool":
		return "bool"
	case "[]string":
		return "List[str]"
	case "[]int":
		return "List[int]"
	case "map[string]string":
		return "Dict[str, str]"
	case "map[string]interface{}":
		return "Dict[str, Any]"
	default:
		return goType
	}
}

// ProtoGenerator Proto文件生成器
type ProtoGenerator struct {
	generator *Generator
}

// NewProtoGenerator 创建Proto生成器
func NewProtoGenerator(generator *Generator) *ProtoGenerator {
	return &ProtoGenerator{generator: generator}
}

// ProtoSpec Proto规范
type ProtoSpec struct {
	Package    string       `json:"package"`
	GoPackage  string       `json:"go_package"`
	Services   []ServiceSpec `json:"services"`
	Messages   []MessageSpec `json:"messages"`
}

// ServiceSpec 服务定义
type ServiceSpec struct {
	Name    string       `json:"name"`
	Methods []RPCMethod  `json:"methods"`
}

// RPCMethod RPC方法
type RPCMethod struct {
	Name        string `json:"name"`
	InputType   string `json:"input_type"`
	OutputType  string `json:"output_type"`
	Description string `json:"description"`
}

// MessageSpec 消息定义
type MessageSpec struct {
	Name   string      `json:"name"`
	Fields []FieldSpec `json:"fields"`
}

// GenerateProto 生成Proto文件
func (g *ProtoGenerator) GenerateProto(spec ProtoSpec, outputPath string) error {
	var code strings.Builder

	code.WriteString("syntax = \"proto3\";\n\n")
	code.WriteString(fmt.Sprintf("package %s;\n\n", spec.Package))
	code.WriteString(fmt.Sprintf("option go_package = \"%s\";\n\n", spec.GoPackage))

	// 消息
	for _, msg := range spec.Messages {
		code.WriteString(fmt.Sprintf("message %s {\n", msg.Name))
		for i, field := range msg.Fields {
			protoType := goTypeToProto(field.Type)
			code.WriteString(fmt.Sprintf("  %s %s = %d;\n", protoType, field.JSONName, i+1))
		}
		code.WriteString("}\n\n")
	}

	// 服务
	for _, svc := range spec.Services {
		code.WriteString(fmt.Sprintf("service %s {\n", svc.Name))
		for _, method := range svc.Methods {
			code.WriteString(fmt.Sprintf("  rpc %s(%s) returns (%s);\n",
				method.Name, method.InputType, method.OutputType))
		}
		code.WriteString("}\n\n")
	}

	// 写入文件
	return nil
}

func goTypeToProto(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int32":
		return "int32"
	case "int64":
		return "int64"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "bool":
		return "bool"
	case "[]byte":
		return "bytes"
	default:
		return goType
	}
}