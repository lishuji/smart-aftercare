package doubao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultBaseURL 豆包 API 默认基础 URL
	DefaultBaseURL = "https://maas-api.ml-platform-cn-beijing.volces.com/api/v3"
	// DefaultTimeout HTTP 请求默认超时时间
	DefaultTimeout = 60 * time.Second
)

// Client 豆包大模型客户端
type Client struct {
	apiKey     string
	chatModel  string
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建豆包客户端
func NewClient(apiKey, chatModel string) *Client {
	return &Client{
		apiKey:    apiKey,
		chatModel: chatModel,
		baseURL:   DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// WithBaseURL 设置自定义 API 基础 URL
func (c *Client) WithBaseURL(baseURL string) *Client {
	c.baseURL = baseURL
	return c
}

// WithTimeout 设置 HTTP 请求超时时间
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

// ==================== 聊天补全 ====================

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatChoice 聊天回复选项
type ChatChoice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string       `json:"id"`
	Choices []ChatChoice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Generate 调用大模型生成回答
func (c *Client) Generate(ctx context.Context, prompt string, temperature float64, maxTokens int) (string, error) {
	req := ChatRequest{
		Model: c.chatModel,
		Messages: []ChatMessage{
			{Role: "system", Content: "你是一个专业的家电售后服务助手，根据用户提供的家电说明书内容，准确、专业地回答用户关于家电操作、故障排查、保养维护等问题。"},
			{Role: "user", Content: prompt},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误(HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("大模型未返回有效回答")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ==================== Embedding ====================

// EmbeddingRequest 向量化请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingData 单条向量化数据
type EmbeddingData struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// EmbeddingResponse 向量化响应
type EmbeddingResponse struct {
	Data  []EmbeddingData `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateEmbeddings 批量生成文本向量
func GenerateEmbeddings(apiKey, embeddingModel string, texts []string) ([][]float32, error) {
	return generateEmbeddingsWithURL(DefaultBaseURL, apiKey, embeddingModel, texts)
}

// GenerateEmbedding 生成单条文本向量
func GenerateEmbedding(apiKey, embeddingModel string, text string) ([]float32, error) {
	vectors, err := GenerateEmbeddings(apiKey, embeddingModel, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("向量化结果为空")
	}
	return vectors[0], nil
}

// generateEmbeddingsWithURL 内部实现：调用向量化 API
func generateEmbeddingsWithURL(baseURL, apiKey, embeddingModel string, texts []string) ([][]float32, error) {
	req := EmbeddingRequest{
		Model: embeddingModel,
		Input: texts,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Embedding API返回错误(HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var embResp EmbeddingResponse
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 按 index 排序结果
	vectors := make([][]float32, len(texts))
	for _, data := range embResp.Data {
		if data.Index < len(vectors) {
			vectors[data.Index] = data.Embedding
		}
	}

	return vectors, nil
}
