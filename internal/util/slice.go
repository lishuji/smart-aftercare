package util

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// TextSlice 文本切片结果
type TextSlice struct {
	Content string
	Index   int
}

// SplitText 按固定长度切片文本（支持重叠）
// chunkSize: 每片最大字符数
// overlap: 相邻切片重叠字符数
func SplitText(text string, chunkSize, overlap int) []string {
	if text == "" {
		return nil
	}

	text = strings.TrimSpace(text)
	textLen := utf8.RuneCountInString(text)

	if textLen <= chunkSize {
		return []string{text}
	}

	runes := []rune(text)
	var slices []string
	start := 0

	for start < textLen {
		end := start + chunkSize
		if end > textLen {
			end = textLen
		}

		chunk := string(runes[start:end])
		chunk = strings.TrimSpace(chunk)
		if chunk != "" {
			slices = append(slices, chunk)
		}

		if end >= textLen {
			break
		}

		// 移动到下一个切片位置（考虑重叠）
		start = end - overlap
		if start <= 0 {
			break
		}
	}

	return slices
}

// SplitByParagraph 按段落切片（优先在段落边界切分）
func SplitByParagraph(text string, maxChunkSize int) []string {
	if text == "" {
		return nil
	}

	// 按换行符分割段落
	paragraphs := strings.Split(text, "\n")
	var slices []string
	var currentChunk strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// 如果当前段落加上累积的内容超过限制，先保存当前内容
		if currentChunk.Len() > 0 && utf8.RuneCountInString(currentChunk.String()+para) > maxChunkSize {
			slices = append(slices, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
		}

		// 如果单个段落就超过限制，按字符切分
		if utf8.RuneCountInString(para) > maxChunkSize {
			subSlices := SplitText(para, maxChunkSize, 30)
			slices = append(slices, subSlices...)
			continue
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n")
		}
		currentChunk.WriteString(para)
	}

	// 保存最后的内容
	if currentChunk.Len() > 0 {
		slices = append(slices, strings.TrimSpace(currentChunk.String()))
	}

	return slices
}

// Chapter 章节信息
type Chapter struct {
	Title    string
	StartPage int
	EndPage   int
	Level     int // 1=一级标题, 2=二级标题, ...
}

// ParseChapters 从文本页面中提取章节结构
func ParseChapters(pages []PageContent) []Chapter {
	var chapters []Chapter

	// 章节匹配正则（支持中英文序号）
	chapterPatterns := []*regexp.Regexp{
		// 匹配 "第X章"、"第X节"
		regexp.MustCompile(`第[一二三四五六七八九十\d]+[章节][\s:：]*(.+)`),
		// 匹配 "1. xxx"、"1.1 xxx"
		regexp.MustCompile(`^(\d+(?:\.\d+)*)[\.、\s]+(.+)`),
		// 匹配 "一、xxx"、"（一）xxx"
		regexp.MustCompile(`^[（(]?[一二三四五六七八九十]+[）)、][\.、\s]*(.+)`),
	}

	for _, page := range pages {
		lines := strings.Split(page.Text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || utf8.RuneCountInString(line) > 50 {
				continue
			}

			for _, pattern := range chapterPatterns {
				if pattern.MatchString(line) {
					level := 1
					// 判断层级
					if strings.Contains(line, ".") {
						dotCount := strings.Count(line, ".")
						level = dotCount + 1
					}

					chapter := Chapter{
						Title:     line,
						StartPage: page.PageNum,
						Level:     level,
					}
					chapters = append(chapters, chapter)
					break
				}
			}
		}
	}

	// 设置章节结束页
	for i := range chapters {
		if i < len(chapters)-1 {
			chapters[i].EndPage = chapters[i+1].StartPage
		} else {
			chapters[i].EndPage = 9999 // 最后一个章节延续到文档末尾
		}
	}

	return chapters
}

// GetCurrentChapter 根据页码获取当前所属章节
func GetCurrentChapter(pageNum int, chapters []Chapter) *Chapter {
	for i := len(chapters) - 1; i >= 0; i-- {
		if pageNum >= chapters[i].StartPage {
			return &chapters[i]
		}
	}
	// 默认返回 "前言" 章节
	return &Chapter{
		Title:     "前言",
		StartPage: 1,
		Level:     1,
	}
}

// PageContent 页面内容
type PageContent struct {
	PageNum int
	Text    string
	Images  [][]byte
}
