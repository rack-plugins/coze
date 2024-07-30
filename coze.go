package coze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fimreal/goutils/ezap"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	apiVersion = "v2"
)

// RequestPayload 定义接收的请求数据结构
type RequestPayload struct {
	UserID         string `json:"user_id" form:"user_id" binding:"required"`
	BotID          string `json:"botid" form:"botid" binding:"required"`
	Prompt         string `json:"prompt" form:"prompt" binding:"required"`
	ConversationID string `json:"conversation_id" form:"conversation_id"`
}

// ResponseContent 定义返回的数据结构
type ResponseContent struct {
	Code           int    `json:"code"`
	Content        string `json:"content"`
	ConversationID string `json:"conversation_id"`
}

// CozeResponse 定义接口返回结构
type CozeResponse struct {
	Messages       []Message `json:"messages"`
	ConversationID string    `json:"conversation_id"`
	Code           int       `json:"code"`
	Msg            string    `json:"msg"`
}

// Message 定义消息结构
type Message struct {
	Role        string `json:"role"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

// handleTxt2Img 处理文本到图像的转换请求
func handleTxt2Img(c *gin.Context) {
	var payload RequestPayload

	// 尝试绑定 JSON 数据或者表单数据
	if err := c.ShouldBindJSON(&payload); err != nil {
		if err := c.ShouldBind(&payload); err != nil {
			ezap.Errorf("Error binding request data: %v", err)
			c.JSON(http.StatusBadRequest, ResponseContent{Code: http.StatusBadRequest, Content: "Invalid input"})
			return
		}
	}

	// 打印请求数据
	ezap.Debug("coze API request: %v", payload)

	apiResponse, err := callChatApi(payload.UserID, payload.BotID, payload.Prompt, payload.ConversationID)
	if err != nil {
		ezap.Errorf("Error calling chat API: %+v", err)
		c.JSON(http.StatusInternalServerError, ResponseContent{Code: http.StatusInternalServerError, Content: "Failed to call chat API"})
		return
	}

	// 打印 coze API 响应
	ezap.Debug("coze API response: %+v", apiResponse)

	answerContent := extractAnswerContent(apiResponse)
	if answerContent == "" {
		c.JSON(http.StatusInternalServerError, ResponseContent{Code: http.StatusInternalServerError, Content: "Sorry, no valid answer found"})
		return
	}

	// 在响应中包含 ConversationID
	c.JSON(http.StatusOK, ResponseContent{
		Code:           http.StatusOK,
		Content:        answerContent,
		ConversationID: apiResponse.ConversationID,
	})
}

// callChatApi 调用外部API并返回结果
func callChatApi(userID, botID, query, conversationID string) (*CozeResponse, error) {
	url := fmt.Sprintf("%s/open_api/%s/chat", viper.GetString("coze_url"), apiVersion)
	token := viper.GetString("coze_token")

	payload := struct {
		ConversationID string `json:"conversation_id,omitempty"` // omitempty 用于在为空时省略该字段
		BotID          string `json:"bot_id"`
		UserID         string `json:"user"`
		Query          string `json:"query"`
		Stream         bool   `json:"stream"`
	}{
		ConversationID: conversationID,
		BotID:          botID,
		UserID:         userID,
		Query:          query,
		Stream:         false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %s", response.Status)
	}

	var cozeResponse CozeResponse
	if err := json.NewDecoder(response.Body).Decode(&cozeResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &cozeResponse, nil
}

// extractAnswerContent 提取 assistant 角色中的 answer 内容
func extractAnswerContent(cozeResponse *CozeResponse) string {
	for _, message := range cozeResponse.Messages {
		if message.Role == "assistant" && message.Type == "answer" {
			return message.Content // 直接返回 Content 字段
		}
	}
	return "" // 如果没有找到，则返回空字符串
}
