package glm

import (
	"bufio"
	"github.com/ssgo/httpclient"
	"github.com/ssgo/u"
	"io"
	"net/http"
	"strings"
	"time"
)

type GlmCfg struct {
	ApiKey string
	Model  string
}

var glmCfg = GlmCfg{}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool 只有type选择的工具有效，多个工具需传多个Tool
// TODO 支持Function与WebSearch的结构体
type Tool struct {
	Type      string      `json:"type"`
	Retrieval interface{} `json:"retrieval"`
	Function  interface{} `json:"function"`
	WebSearch interface{} `json:"web_search"`
}
type Retrieval struct {
	KnowledgeId    string `json:"knowledge_id"`
	PromptTemplate string `json:"prompt_template"`
}
type Function struct {
}
type WebSearch struct {
}

// Data 请求数据
// joinSentences 是否分句
type Data struct {
	Model         string                    `json:"model"`
	Messages      *[]map[string]interface{} `json:"messages"`
	Stream        bool                      `json:"stream"`
	Temperature   float64                   `json:"temperature"`
	Top_p         float64                   `json:"top_p"`
	Tools         []Tool                    `json:"tools"`
	joinSentences bool
	callback      func(string)
}

type choices struct {
	Delta Message
}

// v4SSEResp 流式响应
type v4SSEResp struct {
	Id      string
	Choices []choices
}

// v4SyncResp 同步响应
type v4SyncResp struct {
	Choices []struct {
		Message struct {
			Content string
		}
	}
}

// MakeV4Data 创建请求体
// joinSentences 是否分句，仅在SSESend中有效
// callback 回调函数，仅在SSESend中有效
func MakeV4Data(prompt string, parentMessage *[]map[string]interface{}, callback func(string), joinSentences bool) (data *Data) {
	logger.Info("======MakeV4Data", "pm", parentMessage)
	data = &Data{
		Model:         glmCfg.Model,
		callback:      callback,
		joinSentences: joinSentences,
	}
	if parentMessage != nil {
		data.Messages = parentMessage
	} else {
		*data.Messages = make([]map[string]interface{}, 0)
	}
	*data.Messages = append(*data.Messages, map[string]interface{}{
		"role":    "user",
		"content": prompt,
	})
	return
}

func (d *Data) V4ToMap() (r map[string]interface{}) {
	u.Convert(d, &r)
	return
}

func V4SSERespResolve(reader *bufio.Reader, ifJoinSentences bool, callback func(string)) (meta map[string]interface{}, answers []string, err error) {
	meta = map[string]interface{}{}
	for {
		buf, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return meta, []string{}, err
		}
		line := string(buf)
		linesp := strings.Split(line, ":")
		if len(linesp) <= 1 {
			continue
		}
		if strings.Contains(linesp[1], "[DONE]") {
			break
		}
		realLine := strings.Join(linesp[1:len(linesp)], ":")
		lineMap := u.UnJsonMap(realLine)
		lineStruct := v4SSEResp{}
		u.Convert(lineMap, &lineStruct)
		if len(lineStruct.Choices) < 1 {
			continue
		}
		v := lineStruct.Choices[0].Delta.Content
		if v == "" {
			v = "\n"
		}
		answers = append(answers, v)
		if ifJoinSentences {
			joinSentences(v, callback)
		} else {
			callback(v)
		}
	}
	return
}

// V4SSESend SSE发送请求
func (d *Data) V4SSESend() (meta map[string]interface{}, err error) {
	d.Stream = true
	if d.Top_p == 0 {
		d.Top_p = 0.7
	}
	if d.Temperature == 0 {
		d.Temperature = 0.95
	}
	req := d.V4ToMap()
	reader := SSESendRaw(req, "https://open.bigmodel.cn/api/paas/v4/chat/completions")
	meta, ans, err := V4SSERespResolve(reader, d.joinSentences, d.callback)
	if err != nil {
		return
	}
	*d.Messages = append(*d.Messages, map[string]interface{}{
		"role":    "assistant",
		"content": strings.Join(ans, ""),
	})
	return
}

// V4SyncSend 同步发送
func (d *Data) V4SyncSend() (string, error) {
	rc := httpclient.GetClient(time.Second*30).Post("https://open.bigmodel.cn/api/paas/v4/chat/completions", d, "Authorization", GenerateToken(glmCfg.ApiKey, 5))
	rr := rc.Map()
	r := v4SyncResp{}
	u.Convert(rr, &r)
	if len(r.Choices) < 1 {
		logger.Error("No Choices", "Response", rc.String())
		return rc.String(), http.ErrBodyNotAllowed
	}
	*d.Messages = append(*d.Messages, map[string]interface{}{
		"role":    "assistant",
		"content": r.Choices[0].Message.Content,
	})
	return r.Choices[0].Message.Content, nil
}
