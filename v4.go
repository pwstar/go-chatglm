package glm

import (
	"bufio"
	"github.com/ssgo/httpclient"
	"github.com/ssgo/u"
	"io"
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

// Tool TODO 支持Function与WebSearch的结构体
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

// Data 请求结构体
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

type v4Resp struct {
	Id      string
	Choices []choices
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

// V4SSESend SSE发送请求
func (d *Data) V4SSESend() (meta map[string]interface{}, err error) {
	d.Stream = true
	if d.Top_p == 0 {
		d.Top_p = 0.7
	}
	if d.Temperature == 0 {
		d.Temperature = 0.95
	}
	logger.Info("===== V4SSESend", "Data", d)
	c := httpclient.GetClient(time.Second * 30)
	r := c.ManualDo("POST", "https://open.bigmodel.cn/api/paas/v4/chat/completions", d, "Authorization", GenerateToken(glmCfg.ApiKey, 5))
	reader := bufio.NewReader(r.Response.Body)
	meta = map[string]interface{}{}
	res := make([]string, 0)
	for {
		buf, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return meta, err
		}
		line := string(buf)
		//logger.Info("readline", "line", line)
		linesp := strings.Split(line, ":")
		if len(linesp) <= 1 {
			continue
		}
		if strings.Contains(linesp[1], "[DONE]") {
			break
		}
		realLine := strings.Join(linesp[1:len(linesp)], ":")
		lineMap := u.UnJsonMap(realLine)
		lineStruct := v4Resp{}
		u.Convert(lineMap, &lineStruct)
		//logger.Info("readline", "lineStruct", lineStruct)

		if len(lineStruct.Choices) < 1 {
			continue
		}
		v := lineStruct.Choices[0].Delta.Content
		if v == "" {
			v = "\n"
		}
		res = append(res, v)
		if d.joinSentences {
			joinSentences(v, d.callback)
		} else {
			d.callback(v)
		}
	}
	*d.Messages = append(*d.Messages, map[string]interface{}{
		"role":    "assistant",
		"content": strings.Join(res, ""),
	})
	return
}

//TODO 同步发送
//func SyncSend(tIn string) string {
//	Data := Data{
//		Stream:      false,
//		Model:       glmCfg.Model,
//		Messages:    make([]map[string]interface{}, 0),
//		Temperature: glmCfg.Temperature,
//		Top_p:       glmCfg.Top_p,
//		Tool: []Tool{
//			{
//				Type: "retrieval",
//				Retrieval: retrieval{
//					Knowledge_id:    glmCfg.Knowledge_id,
//					Prompt_template: glmCfg.Prompt_template,
//				},
//			},
//		},
//	}
//	msg := make(map[string]interface{})
//	log.DefaultLogger.Info("====", "msg", msg)
//	Data.Messages = append(Data.Messages, map[string]interface{}{
//		"role":    "user",
//		"content": tIn,
//	})
//	d := u.Json(Data)
//	log.DefaultLogger.Info("====", "Data", Data)
//	tk := GenerateToken(glmCfg.ApiKey, 10)
//	respraw := httpclient.GetClient(time.Second*30).Post("https://open.bigmodel.cn/api/paas/v4/chat/completions", d, "Authorization", tk, "Content-Type", "application/json").Map()
//	log.DefaultLogger.Info("=====Ans", "respraw", respraw)
//
//	Resp := v4Resp{}
//	u.Convert(respraw, &Resp)
//	log.DefaultLogger.Info("====", "resp", Resp, "ansText", Resp.Choices[0].Message.Content)
//
//	return Resp.Choices[0].Message.Content
//}
