//It's out of date.Don't use it.

package glm

import (
	"bufio"
	"fmt"
	"github.com/ssgo/httpclient"
	"github.com/ssgo/log"
	"github.com/ssgo/u"
	"io"
	"strings"
	"time"
)

type InvokeRequest struct {
	Model  string                   `json:"model"`
	Prompt []map[string]interface{} `json:"prompt"`
}

func SSESend(apiKey string, m string, pm []map[string]interface{}, callback func(string)) (meta map[string]interface{}, err error) {
	token := GenerateToken(apiKey, 300)
	log.DefaultLogger.Info("======", "token", token)
	msg := pm
	msg = append(msg, map[string]interface{}{"role": "user", "content": m})
	request := InvokeRequest{
		Model:  "glm-4",
		Prompt: msg,
	}
	log.DefaultLogger.Info("========", "request", request)
	c := httpclient.GetClient(time.Second * 30)
	r := c.ManualDo("POST", "https://open.bigmodel.cn/api/paas/v3/model-api/glm-3-turbo/"+"sse-invoke", request, "Authorization", token, "Accept", "text/event-stream")
	fmt.Println("=====sent")
	reader := bufio.NewReader(r.Response.Body)
	lastEvent := ""
	log.DefaultLogger.Info("=======read", "reader", reader)
	//lastId := ""
	meta = map[string]interface{}{}
	for {
		buf, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return meta, err
		}
		line := string(buf)
		pos := strings.IndexByte(line, ':')
		if pos != -1 {
			k := line[0:pos]
			v := line[pos+1:]
			switch k {
			case "id":
				//lastId = v
			case "event":
				lastEvent = v
				log.DefaultLogger.Info("=====", "lastEvent", lastEvent)
			case "Data":
				if v == "" {
					v = "\n"
				}
				callback(v)
			case "meta":
				meta = u.UnJsonMap(v)
			}
		}
	}
	return meta, nil
}
