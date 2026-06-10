package proxy

import (
	"ai-gateway/internal/router"
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
	client *http.Client
	router *router.Router
}

func NewHandler(r *router.Router) *Handler {
	return &Handler{
		client: &http.Client{Timeout: 60 * time.Second},
		router: r,
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
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	// 解析失败就当做 false
	_ = json.Unmarshal(body, &meta)

	// 查看是否选择了模型
	model, cb, err := h.router.Get(meta.Model)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var resp *http.Response
	for cb.Allow() {
		// 构建上游请求
		req, err := buildRequest(ctx, body, model)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 转发响应, 等待响应
		resp, err = h.client.Do(req)
		if err != nil {
			cb.ReportFailure()
			resp = nil
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			cb.ReportFailure()
			resp = nil
			time.Sleep(200 * time.Millisecond)
			continue
		}

		cb.ReportSuccess()
		break
	}

	if resp == nil {
		ctx.JSON(http.StatusServiceUnavailable,
			gin.H{"error": fmt.Sprintf("模型 %s 当前不可用，请切换其他模型", meta.Model)})
		return
	}
	defer resp.Body.Close()
	forwardResponse(ctx, resp, meta.Stream)
}

// 构建上游请求
// 将 body 中的 model 字段替换为上游真实模型名，屏蔽网关内部命名
func buildRequest(ctx *gin.Context, body []byte, model config.ModelConfig) (*http.Request, error) {
	var bodyMap map[string]any
	if err := json.Unmarshal(body, &bodyMap); err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}
	bodyMap["model"] = model.Model
	newBody, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	targetURL := model.BaseURL + ctx.Request.URL.Path
	req, err := http.NewRequestWithContext(ctx.Request.Context(),
		ctx.Request.Method, targetURL, bytes.NewReader(newBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	copyHeaders(req.Header, ctx.Request.Header)
	req.Header.Set("Authorization", "Bearer "+model.APIKey)
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
