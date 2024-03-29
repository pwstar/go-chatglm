package main

import (
	"github.com/ssgo/log"
	"glm"
)

func V4SSE() {
	pm := make([]map[string]interface{}, 0)
	glm.SetCfg(glm.GlmCfg{
		Model:  "glm-4",
		ApiKey: "",
	})
	data := glm.MakeV4Data("你好", &pm, func(s string) {
		log.DefaultLogger.Info("CallBack", "str", s)
	}, true)
	data.Top_p = 0.8
	data.Temperature = 0.7
	data.Tools = make([]glm.Tool, 0)
	data.Tools[0] = glm.Tool{
		Type: "retrieval",
		Retrieval: glm.Retrieval{
			KnowledgeId:    "",
			PromptTemplate: "",
		},
	}
	//其他tools类型未完美支持
	mt, err := data.V4SSESend()
	if err != nil {
		log.DefaultLogger.Error("V4 SSE Send Error", "err", err, "meta", mt)
	}
}
