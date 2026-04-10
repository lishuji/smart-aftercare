// Package docs Smart Aftercare API 文档
package docs

import (
	_ "embed"

	"github.com/swaggo/swag"
)

//go:embed swagger.json
var swaggerJSON string

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8000",
	BasePath:         "/api",
	Schemes:          []string{},
	Title:            "智能售后服务系统 API",
	Description:      "基于 RAG 的智能家电售后问答与文档管理系统。",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  swaggerJSON,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
