// Package openapi - OpenAPI文档生成器
// Sprint 25: 正式发布准备 - API文档
package openapi

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// OpenAPIDoc OpenAPI文档
type OpenAPIDoc struct {
	OpenAPI    string                 `json:"openapi"`
	Info       Info                   `json:"info"`
	Servers    []Server               `json:"servers,omitempty"`
	Paths      map[string]PathItem    `json:"paths"`
	Components *Components            `json:"components,omitempty"`
	Tags       []Tag                  `json:"tags,omitempty"`
}

// Info API信息
type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	Version        string   `json:"version"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
}

// Contact 联系信息
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License 许可信息
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server 服务器配置
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable 服务器变量
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem 路径项
type PathItem struct {
	Ref        string     `json:"$ref,omitempty"`
	Summary    string     `json:"summary,omitempty"`
	Description string    `json:"description,omitempty"`
	Get        *Operation `json:"get,omitempty"`
	Put        *Operation `json:"put,omitempty"`
	Post       *Operation `json:"post,omitempty"`
	Delete     *Operation `json:"delete,omitempty"`
	Options    *Operation `json:"options,omitempty"`
	Head       *Operation `json:"head,omitempty"`
	Patch      *Operation `json:"patch,omitempty"`
	Trace      *Operation `json:"trace,omitempty"`
}

// Operation 操作
type Operation struct {
	Tags           []string               `json:"tags,omitempty"`
	Summary        string                 `json:"summary,omitempty"`
	Description    string                 `json:"description,omitempty"`
	OperationID    string                 `json:"operationId"`
	Parameters     []Parameter            `json:"parameters,omitempty"`
	RequestBody    *RequestBody           `json:"requestBody,omitempty"`
	Responses      map[string]Response    `json:"responses"`
	Callbacks      map[string]Callback    `json:"callbacks,omitempty"`
	Deprecated     bool                   `json:"deprecated,omitempty"`
	Security       []SecurityRequirement  `json:"security,omitempty"`
	Servers        []Server               `json:"servers,omitempty"`
	ExternalDocs   *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// Parameter 参数
type Parameter struct {
	Name            string      `json:"name"`
	In              string      `json:"in"` // query, header, path, cookie
	Description     string      `json:"description,omitempty"`
	Required        bool        `json:"required"`
	Deprecated      bool        `json:"deprecated,omitempty"`
	Schema          *Schema     `json:"schema,omitempty"`
	Example         interface{} `json:"example,omitempty"`
	Examples        map[string]Example `json:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty"`
}

// RequestBody 请求体
type RequestBody struct {
	Description string                `json:"description,omitempty"`
	Content     map[string]MediaType  `json:"content"`
	Required    bool                  `json:"required,omitempty"`
}

// MediaType 媒体类型
type MediaType struct {
	Schema    *Schema             `json:"schema,omitempty"`
	Example   interface{}         `json:"example,omitempty"`
	Examples  map[string]Example  `json:"examples,omitempty"`
	Encoding  map[string]Encoding `json:"encoding,omitempty"`
}

// Schema 模式
type Schema struct {
	Ref                  string             `json:"$ref,omitempty"`
	Type                 string             `json:"type,omitempty"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty"`
	Format               string             `json:"format,omitempty"`
	Default              interface{}        `json:"default,omitempty"`
	Example              interface{}        `json:"example,omitempty"`
	Nullable             bool               `json:"nullable,omitempty"`
	ReadOnly             bool               `json:"readOnly,omitempty"`
	WriteOnly            bool               `json:"writeOnly,omitempty"`
	AdditionalProperties interface{}        `json:"additionalProperties,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty"`
}

// Response 响应
type Response struct {
	Description string                `json:"description"`
	Headers     map[string]Header     `json:"headers,omitempty"`
	Content     map[string]MediaType  `json:"content,omitempty"`
	Links       map[string]Link       `json:"links,omitempty"`
}

// Header 头
type Header struct {
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Deprecated  bool    `json:"deprecated,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// Example 示例
type Example struct {
	Summary       string      `json:"summary,omitempty"`
	Description   string      `json:"description,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}

// Encoding 编码
type Encoding struct {
	ContentType   string               `json:"contentType,omitempty"`
	Headers       map[string]Header    `json:"headers,omitempty"`
	Style         string               `json:"style,omitempty"`
	Explode       bool                 `json:"explode,omitempty"`
	AllowReserved bool                 `json:"allowReserved,omitempty"`
}

// Link 链接
type Link struct {
	OperationRef string                 `json:"operationRef,omitempty"`
	OperationID  string                 `json:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Server       *Server                `json:"server,omitempty"`
}

// Callback 回调
type map[string]PathItem

// SecurityRequirement 安全要求
type SecurityRequirement map[string][]string

// Components 组件
type Components struct {
	Schemas         map[string]*Schema        `json:"schemas,omitempty"`
	Responses       map[string]Response       `json:"responses,omitempty"`
	Parameters      map[string]Parameter      `json:"parameters,omitempty"`
	Examples        map[string]Example        `json:"examples,omitempty"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies,omitempty"`
	Headers         map[string]Header         `json:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
	Links           map[string]Link           `json:"links,omitempty"`
	Callbacks       map[string]Callback       `json:"callbacks,omitempty"`
}

// SecurityScheme 安全方案
type SecurityScheme struct {
	Type             string `json:"type"`
	Description      string `json:"description,omitempty"`
	Name             string `json:"name,omitempty"`
	In               string `json:"in,omitempty"`
	Scheme           string `json:"scheme,omitempty"`
	BearerFormat     string `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIdConnectUrl string `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows OAuth流程
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow OAuth流程
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// Tag 标签
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// ExternalDocumentation 外部文档
type ExternalDocumentation struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// DocumentGenerator 文档生成器
type DocumentGenerator struct {
	doc *OpenAPIDoc
}

// NewDocumentGenerator 创建文档生成器
func NewDocumentGenerator(title, version string) *DocumentGenerator {
	return &DocumentGenerator{
		doc: &OpenAPIDoc{
			OpenAPI: "3.0.3",
			Info: Info{
				Title:       title,
				Version:     version,
				Description: "OFA - Omni Federated Agents API",
			},
			Paths: make(map[string]PathItem),
		},
	}
}

// SetDescription 设置描述
func (dg *DocumentGenerator) SetDescription(desc string) *DocumentGenerator {
	dg.doc.Info.Description = desc
	return dg
}

// SetContact 设置联系信息
func (dg *DocumentGenerator) SetContact(name, email, url string) *DocumentGenerator {
	dg.doc.Info.Contact = &Contact{
		Name:  name,
		Email: email,
		URL:   url,
	}
	return dg
}

// SetLicense 设置许可
func (dg *DocumentGenerator) SetLicense(name, url string) *DocumentGenerator {
	dg.doc.Info.License = &License{
		Name: name,
		URL:  url,
	}
	return dg
}

// AddServer 添加服务器
func (dg *DocumentGenerator) AddServer(url, desc string) *DocumentGenerator {
	dg.doc.Servers = append(dg.doc.Servers, Server{
		URL:         url,
		Description: desc,
	})
	return dg
}

// AddTag 添加标签
func (dg *DocumentGenerator) AddTag(name, desc string) *DocumentGenerator {
	dg.doc.Tags = append(dg.doc.Tags, Tag{
		Name:        name,
		Description: desc,
	})
	return dg
}

// AddPath 添加路径
func (dg *DocumentGenerator) AddPath(path string, item PathItem) *DocumentGenerator {
	dg.doc.Paths[path] = item
	return dg
}

// AddGET 添加GET操作
func (dg *DocumentGenerator) AddGET(path string, op *Operation) *DocumentGenerator {
	if _, ok := dg.doc.Paths[path]; !ok {
		dg.doc.Paths[path] = PathItem{}
	}
	item := dg.doc.Paths[path]
	item.Get = op
	dg.doc.Paths[path] = item
	return dg
}

// AddPOST 添加POST操作
func (dg *DocumentGenerator) AddPOST(path string, op *Operation) *DocumentGenerator {
	if _, ok := dg.doc.Paths[path]; !ok {
		dg.doc.Paths[path] = PathItem{}
	}
	item := dg.doc.Paths[path]
	item.Post = op
	dg.doc.Paths[path] = item
	return dg
}

// AddDELETE 添加DELETE操作
func (dg *DocumentGenerator) AddDELETE(path string, op *Operation) *DocumentGenerator {
	if _, ok := dg.doc.Paths[path]; !ok {
		dg.doc.Paths[path] = PathItem{}
	}
	item := dg.doc.Paths[path]
	item.Delete = op
	dg.doc.Paths[path] = item
	return dg
}

// SetComponents 设置组件
func (dg *DocumentGenerator) SetComponents(components *Components) *DocumentGenerator {
	dg.doc.Components = components
	return dg
}

// GenerateOFAAPI 生成OFA API文档
func GenerateOFAAPI() *OpenAPIDoc {
	gen := NewDocumentGenerator("OFA API", "v7.1.0")
	gen.SetDescription("OFA - Omni Federated Agents 分布式智能体系统 API").
		SetContact("OFA Team", "ofa@example.com", "https://ofa.dev").
		SetLicense("MIT", "https://opensource.org/licenses/MIT").
		AddServer("http://localhost:8080", "本地开发服务器").
		AddServer("https://api.ofa.dev", "生产服务器").
		AddTag("agent", "Agent管理相关接口").
		AddTag("task", "任务管理相关接口").
		AddTag("skill", "技能管理相关接口").
		AddTag("message", "消息通信相关接口").
		AddTag("system", "系统管理相关接口")

	// 健康检查
	gen.AddGET("/health", &Operation{
		Tags:        []string{"system"},
		Summary:     "健康检查",
		Description: "检查服务健康状态",
		OperationID: "healthCheck",
		Responses: map[string]Response{
			"200": {
				Description: "服务正常",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]*Schema{
								"status": {Type: "string", Example: "ok"},
							},
						},
					},
				},
			},
		},
	})

	// Agent列表
	gen.AddGET("/api/v1/agents", &Operation{
		Tags:        []string{"agent"},
		Summary:     "获取Agent列表",
		Description: "获取所有注册的Agent列表",
		OperationID: "listAgents",
		Responses: map[string]Response{
			"200": {
				Description: "Agent列表",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{
							Type: "array",
							Items: &Schema{Ref: "#/components/schemas/Agent"},
						},
					},
				},
			},
		},
	})

	// Agent注册
	gen.AddPOST("/api/v1/agents", &Operation{
		Tags:        []string{"agent"},
		Summary:     "注册Agent",
		Description: "注册新的Agent到系统",
		OperationID: "registerAgent",
		RequestBody: &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{Ref: "#/components/schemas/AgentRegistration"},
				},
			},
		},
		Responses: map[string]Response{
			"201": {
				Description: "注册成功",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{Ref: "#/components/schemas/Agent"},
					},
				},
			},
		},
	})

	// 任务提交
	gen.AddPOST("/api/v1/tasks", &Operation{
		Tags:        []string{"task"},
		Summary:     "提交任务",
		Description: "提交新任务到系统执行",
		OperationID: "submitTask",
		RequestBody: &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{Ref: "#/components/schemas/TaskRequest"},
				},
			},
		},
		Responses: map[string]Response{
			"202": {
				Description: "任务已接受",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{Ref: "#/components/schemas/Task"},
					},
				},
			},
		},
	})

	// 任务查询
	gen.AddGET("/api/v1/tasks/{id}", &Operation{
		Tags:        []string{"task"},
		Summary:     "查询任务状态",
		Description: "根据ID查询任务执行状态",
		OperationID: "getTask",
		Parameters: []Parameter{
			{
				Name:     "id",
				In:       "path",
				Required: true,
				Schema:   &Schema{Type: "string"},
			},
		},
		Responses: map[string]Response{
			"200": {
				Description: "任务详情",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{Ref: "#/components/schemas/Task"},
					},
				},
			},
			"404": {
				Description: "任务不存在",
			},
		},
	})

	// 技能列表
	gen.AddGET("/api/v1/skills", &Operation{
		Tags:        []string{"skill"},
		Summary:     "获取技能列表",
		Description: "获取所有可用技能",
		OperationID: "listSkills",
		Responses: map[string]Response{
			"200": {
				Description: "技能列表",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{
							Type: "array",
							Items: &Schema{Ref: "#/components/schemas/Skill"},
						},
					},
				},
			},
		},
	})

	// 系统信息
	gen.AddGET("/api/v1/system/info", &Operation{
		Tags:        []string{"system"},
		Summary:     "获取系统信息",
		Description: "获取系统运行状态信息",
		OperationID: "getSystemInfo",
		Responses: map[string]Response{
			"200": {
				Description: "系统信息",
				Content: map[string]MediaType{
					"application/json": {
						Schema: &Schema{Ref: "#/components/schemas/SystemInfo"},
					},
				},
			},
		},
	})

	// 组件定义
	gen.SetComponents(&Components{
		Schemas: map[string]*Schema{
			"Agent": {
				Type: "object",
				Properties: map[string]*Schema{
					"id":         {Type: "string"},
					"name":       {Type: "string"},
					"type":       {Type: "string"},
					"status":     {Type: "string", Enum: []interface{}{"online", "offline", "busy"}},
					"capabilities": {Type: "array", Items: &Schema{Type: "string"}},
					"last_seen":  {Type: "string", Format: "date-time"},
				},
			},
			"AgentRegistration": {
				Type: "object",
				Required: []string{"name", "type"},
				Properties: map[string]*Schema{
					"name":        {Type: "string"},
					"type":        {Type: "string"},
					"capabilities": {Type: "array", Items: &Schema{Type: "string"}},
					"metadata":    {Type: "object", AdditionalProperties: &Schema{Type: "string"}},
				},
			},
			"Task": {
				Type: "object",
				Properties: map[string]*Schema{
					"id":          {Type: "string"},
					"skill_id":    {Type: "string"},
					"status":      {Type: "string", Enum: []interface{}{"pending", "running", "completed", "failed"}},
					"input":       {Type: "object"},
					"output":      {Type: "object"},
					"error":       {Type: "string"},
					"created_at":  {Type: "string", Format: "date-time"},
					"completed_at": {Type: "string", Format: "date-time"},
				},
			},
			"TaskRequest": {
				Type: "object",
				Required: []string{"skill_id", "input"},
				Properties: map[string]*Schema{
					"skill_id":       {Type: "string"},
					"input":          {Type: "object"},
					"agent_id":       {Type: "string"},
					"priority":       {Type: "integer"},
					"timeout_seconds": {Type: "integer"},
				},
			},
			"Skill": {
				Type: "object",
				Properties: map[string]*Schema{
					"id":          {Type: "string"},
					"name":        {Type: "string"},
					"description": {Type: "string"},
					"operations":  {Type: "array", Items: &Schema{Type: "string"}},
				},
			},
			"SystemInfo": {
				Type: "object",
				Properties: map[string]*Schema{
					"version":     {Type: "string"},
					"uptime":      {Type: "integer"},
					"agents_count": {Type: "integer"},
					"tasks_count": {Type: "integer"},
					"go_version":  {Type: "string"},
				},
			},
		},
		SecuritySchemes: map[string]SecurityScheme{
			"bearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		},
	})

	return gen.doc
}

// ToJSON 导出为JSON
func (doc *OpenAPIDoc) ToJSON() ([]byte, error) {
	return json.MarshalIndent(doc, "", "  ")
}

// ToYAML 导出为YAML（简化版）
func (doc *OpenAPIDoc) ToYAML() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("openapi: %s\n", doc.OpenAPI))
	sb.WriteString("info:\n")
	sb.WriteString(fmt.Sprintf("  title: %s\n", doc.Info.Title))
	sb.WriteString(fmt.Sprintf("  version: %s\n", doc.Info.Version))
	if doc.Info.Description != "" {
		sb.WriteString(fmt.Sprintf("  description: %s\n", doc.Info.Description))
	}
	sb.WriteString("\npaths:\n")
	for path, item := range doc.Paths {
		sb.WriteString(fmt.Sprintf("  %s:\n", path))
		if item.Get != nil {
			sb.WriteString(fmt.Sprintf("    get:\n"))
			sb.WriteString(fmt.Sprintf("      summary: %s\n", item.Get.Summary))
			sb.WriteString(fmt.Sprintf("      operationId: %s\n", item.Get.OperationID))
		}
		if item.Post != nil {
			sb.WriteString(fmt.Sprintf("    post:\n"))
			sb.WriteString(fmt.Sprintf("      summary: %s\n", item.Post.Summary))
			sb.WriteString(fmt.Sprintf("      operationId: %s\n", item.Post.OperationID))
		}
	}
	return sb.String()
}