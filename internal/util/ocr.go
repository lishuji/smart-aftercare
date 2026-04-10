package util

// NOTE: OCR 功能依赖 Tesseract 系统库。
// 在 Docker 环境中通过 apk add tesseract-ocr tesseract-ocr-data-chi-sim 安装。
// 本地开发时需要安装 Tesseract: brew install tesseract (macOS) 或 apt install tesseract-ocr (Linux)
//
// 如果 gosseract 不可用，以下函数提供降级实现。

import (
	"smart-aftercare/pkg/logger"
	"strings"
)

// ExtractTextWithOCR 提取页面文本（含 OCR 处理图片中的文字）
// 当前为降级实现，直接返回已提取的文本
// 完整实现需要引入 gosseract 库：
//
//	import "github.com/otiai10/gosseract/v2"
//	client := gosseract.NewClient()
//	defer client.Close()
//	client.SetLanguage("chi_sim", "eng")
//	client.SetImageFromBytes(imageData)
//	text, _ := client.Text()
func ExtractTextWithOCR(pageText string, images [][]byte) string {
	var builder strings.Builder
	builder.WriteString(pageText)

	// 如果有图片且需要 OCR，尝试调用 Tesseract
	if len(images) > 0 {
		for _, imgData := range images {
			ocrText := performOCR(imgData)
			if ocrText != "" {
				builder.WriteString("\n")
				builder.WriteString(ocrText)
			}
		}
	}

	return builder.String()
}

// performOCR 执行 OCR 识别
// 当前为降级实现（返回空字符串）
// 完整实现请取消注释下方代码并引入 gosseract
func performOCR(imageData []byte) string {
	if len(imageData) == 0 {
		return ""
	}

	// TODO: 启用完整 OCR 功能
	// 取消以下注释并在 go.mod 中添加 github.com/otiai10/gosseract/v2
	/*
		client := gosseract.NewClient()
		defer client.Close()

		// 设置识别语言（简体中文 + 英文）
		client.SetLanguage("chi_sim", "eng")

		// 设置图片数据
		if err := client.SetImageFromBytes(imageData); err != nil {
			logger.Warnf("设置OCR图片失败: %v", err)
			return ""
		}

		// 执行识别
		text, err := client.Text()
		if err != nil {
			logger.Warnf("OCR识别失败: %v", err)
			return ""
		}

		return strings.TrimSpace(text)
	*/

	logger.Debug("OCR功能未启用，跳过图片文字识别")
	return ""
}

// ExtractImagesFromPDFPage 从 PDF 页面提取图片数据
// 当前为占位实现，完整实现需要配合 PDF 解析库
func ExtractImagesFromPDFPage(pageData interface{}) [][]byte {
	// TODO: 实现 PDF 页面图片提取
	// 使用 unipdf 库时：
	// extractor, _ := pdfPage.GetContentStreamProcessor()
	// 提取 XObject 中的 Image 类型对象
	return nil
}
