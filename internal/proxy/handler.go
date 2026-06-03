package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"

	"ai-gateway/internal/config"
)

type Handler struct {
	client   *http.Client
	upstream config.ModelConfig
}

func NewHandler(upstream config.ModelConfig) *Handler {
	return &Handler{
		client:   &http.Client{Timeout: 60 * time.Second},
		upstream: upstream,
	}
}

func (h *Handler) Handle(ctx *gin.Context) {
	// 读请求体，顺便解析 stream 标志
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var meta struct {
		Stream bool `json:"stream"`
	}
	// 解析失败就当做 false
	_ = json.Unmarshal(body, &meta)

	// 构建上游请求
	req, err := h.buildRequest(ctx, body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("build request: %v", err)})
		return
	}
	// 转发响应, 等待响应
	resp, err := h.client.Do(req)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("upstream: %v", err)})
		return
	}
	defer resp.Body.Close()
	forwardResponse(ctx, resp, meta.Stream)
}

// 构建上游请求
func (h *Handler) buildRequest(ctx *gin.Context, body []byte) (*http.Request, error) {
	targetURL := h.upstream.BaseURL + ctx.Request.URL.Path
	req, err := http.NewRequestWithContext(ctx.Request.Context(),
		ctx.Request.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	copyHeaders(req.Header, ctx.Request.Header)
	req.Header.Set("Authorization", "Bearer "+h.upstream.APIKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// 转发响应给客户端
func forwardResponse(ctx *gin.Context, resp *http.Response, stream bool) {
	for k, vs := range resp.Header {
		for _, v := range vs {
			ctx.Header(k, v)
		}
	}
	ctx.Status(resp.StatusCode)

	if stream {
		flusher, ok := ctx.Writer.(http.Flusher)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
			return
		}
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				_, _ = ctx.Writer.Write(buf[:n])
				flusher.Flush()
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
		}
	} else {
		_, _ = io.Copy(ctx.Writer, resp.Body)
	}
}

// 复制 Header
func copyHeaders(dst, src http.Header) {
	// dst 新请求; src 原始请求
	for k, vs := range src {
		for _, v := range vs {
			dst.Set(k, v)
		}
	}
}
