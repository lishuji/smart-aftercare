package util

import (
	"fmt"
	"strings"
)

// GenerateAppliancePrompt 生成家电场景专用 Prompt
func GenerateAppliancePrompt(query, contextText, modelName string) string {
	prompt := fmt.Sprintf(`你是一个专业的家电售后服务助手。请根据以下说明书内容，准确回答用户的问题。

## 回答要求
1. 基于提供的说明书内容回答，不要编造信息
2. 如果涉及操作步骤，请按顺序列出
3. 如果涉及故障排查，请先说明可能原因，再给出解决方案
4. 如果提供的内容无法回答问题，请说明"当前资料暂未包含相关信息"
5. 语言简洁、专业，适合普通用户理解

## 家电型号
%s

## 说明书相关内容
%s

## 用户问题
%s

请根据以上信息回答用户的问题：`, modelName, contextText, query)

	return prompt
}

// GenerateErrorCodePrompt 生成故障代码查询 Prompt
func GenerateErrorCodePrompt(code, contextText, modelName string) string {
	prompt := fmt.Sprintf(`你是一个专业的家电故障诊断助手。请根据以下说明书内容，解释故障代码的含义并给出解决方案。

## 回答要求
1. 首先说明故障代码的含义
2. 列出可能的故障原因（按概率从高到低）
3. 给出具体的解决步骤
4. 如果需要专业维修，请提醒用户联系售后
5. 语言简洁、专业

## 家电型号
%s

## 相关资料
%s

## 故障代码
%s

请分析此故障代码：`, modelName, contextText, code)

	return prompt
}

// ExtractApplianceKeywords 从查询中提取家电相关关键词
func ExtractApplianceKeywords(query string) []string {
	var keywords []string

	// 常见家电操作关键词
	operationKeywords := []string{
		"开机", "关机", "启动", "停止", "暂停",
		"制冷", "制热", "除湿", "送风", "自动",
		"定时", "预约", "睡眠", "节能", "静音",
		"温度", "风速", "风向", "摆风",
		"清洗", "保养", "维护", "清洁", "消毒",
		"安装", "拆卸", "移机", "加氟",
		"遥控器", "面板", "显示屏", "指示灯",
		"滤网", "蒸发器", "冷凝器", "压缩机",
		"排水", "进水", "出水", "漏水",
		"噪音", "异响", "震动", "振动",
		"不制冷", "不制热", "不出风", "不启动",
		"漏电", "跳闸", "短路",
		"wifi", "智能", "APP", "联网",
	}

	// 故障代码模式匹配
	queryLower := strings.ToLower(query)
	for _, kw := range operationKeywords {
		if strings.Contains(query, kw) || strings.Contains(queryLower, strings.ToLower(kw)) {
			keywords = append(keywords, kw)
		}
	}

	// 提取故障代码（如 E1, F2, P3, H6 等）
	errorCodePatterns := []string{"E", "F", "P", "H", "L", "U", "C"}
	for _, prefix := range errorCodePatterns {
		for i := 0; i <= 99; i++ {
			code := fmt.Sprintf("%s%d", prefix, i)
			if strings.Contains(strings.ToUpper(query), code) {
				keywords = append(keywords, code)
			}
		}
	}

	// 如果没有提取到关键词，使用分词（简单分词：按空格和标点分割）
	if len(keywords) == 0 {
		words := splitQueryWords(query)
		for _, w := range words {
			if len(w) >= 2 { // 过滤单字
				keywords = append(keywords, w)
			}
		}
	}

	// 去重
	return uniqueStrings(keywords)
}

// splitQueryWords 简单分词（按空格和标点分割）
func splitQueryWords(query string) []string {
	// 替换常见标点为空格
	replacer := strings.NewReplacer(
		"，", " ", "。", " ", "？", " ", "！", " ",
		"、", " ", "；", " ", "：", " ",
		",", " ", ".", " ", "?", " ", "!", " ",
		"(", " ", ")", " ", "（", " ", "）", " ",
	)
	query = replacer.Replace(query)

	words := strings.Fields(query)
	var result []string
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w != "" {
			result = append(result, w)
		}
	}
	return result
}

// uniqueStrings 字符串去重
func uniqueStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
