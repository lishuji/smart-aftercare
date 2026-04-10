package util

import (
	"smart-aftercare/internal/repository"
	"sort"
	"strings"
)

// 章节优先级（故障排查 > 操作指南 > 保养维护 > 其他）
var chapterPriority = map[string]int{
	"故障":   1,
	"排查":   1,
	"错误":   1,
	"维修":   1,
	"报警":   1,
	"代码":   1,
	"操作":   2,
	"使用":   2,
	"功能":   2,
	"指南":   2,
	"说明":   2,
	"保养":   3,
	"维护":   3,
	"清洁":   3,
	"安装":   4,
	"规格":   5,
	"参数":   5,
}

// MergeAndRankResults 合并关键词检索和向量检索结果，并按章节优先级排序
func MergeAndRankResults(keywordResults, vectorResults []*repository.VectorSlice) []*repository.VectorSlice {
	// 使用 map 去重（基于内容的前 100 字符）
	seen := make(map[string]bool)
	var merged []*repository.VectorSlice

	// 关键词结果优先
	for _, r := range keywordResults {
		key := truncateContent(r.Content, 100)
		if !seen[key] {
			seen[key] = true
			merged = append(merged, r)
		}
	}

	// 然后添加向量结果
	for _, r := range vectorResults {
		key := truncateContent(r.Content, 100)
		if !seen[key] {
			seen[key] = true
			merged = append(merged, r)
		}
	}

	// 按章节优先级排序
	sort.SliceStable(merged, func(i, j int) bool {
		pi := getChapterPriority(merged[i].Metadata["chapter"])
		pj := getChapterPriority(merged[j].Metadata["chapter"])
		return pi < pj
	})

	return merged
}

// getChapterPriority 获取章节优先级分数（数字越小优先级越高）
func getChapterPriority(chapterTitle string) int {
	minPriority := 99
	for keyword, priority := range chapterPriority {
		if strings.Contains(chapterTitle, keyword) && priority < minPriority {
			minPriority = priority
		}
	}
	return minPriority
}

// truncateContent 截断内容用于去重比较
func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return content
}

// FormatSources 格式化来源信息
func FormatSources(results []*repository.VectorSlice) []string {
	var sources []string
	seen := make(map[string]bool)

	for _, r := range results {
		source := r.Metadata["brand"] + " " + r.Metadata["model"] +
			"（第" + r.Metadata["page"] + "页，" + r.Metadata["chapter"] + "）"
		if !seen[source] {
			seen[source] = true
			sources = append(sources, source)
		}
	}

	return sources
}

// CollectImageURLs 收集结果中的图片 URL
func CollectImageURLs(results []*repository.VectorSlice) []string {
	var images []string
	seen := make(map[string]bool)

	for _, r := range results {
		if imgURL, ok := r.Metadata["image_url"]; ok && imgURL != "" {
			if !seen[imgURL] {
				seen[imgURL] = true
				images = append(images, imgURL)
			}
		}
	}

	return images
}

// BuildContextText 从检索结果构建上下文文本
func BuildContextText(results []*repository.VectorSlice) string {
	var builder strings.Builder
	for i, r := range results {
		if i > 0 {
			builder.WriteString("\n---\n")
		}
		if chapter, ok := r.Metadata["chapter"]; ok && chapter != "" {
			builder.WriteString("[" + chapter + "] ")
		}
		builder.WriteString(r.Content)
	}
	return builder.String()
}
