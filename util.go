package glm

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ssgo/log"
	"github.com/ssgo/u"
	"os"
	"strings"
	"time"
)

// Punctuations 分句的判断标点
var Punctuations = []string{
	"。",
	"？",
	"！",
	".",
	"?",
	"!",
	"~",
}
var modTime = time.Now()
var Results []string
var Resultsn []string
var logger = log.DefaultLogger

// UploadCfg 更新配置
// path 文件路径
func UploadCfg(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		logger.Error("readModTimeError", "err", err, "info", info)
		return err
	}
	if info.ModTime() != modTime {
		modTime = info.ModTime()
		err = u.LoadYaml(path, &glmCfg)
		if err != nil {
			logger.Error("updateCfgErr", "err", err, "config", glmCfg)
			return err
		}
		logger.Info("uploadCfg", "config", glmCfg)
	}
	return nil
}

// SetCfg 更新配置
// config 配置
func SetCfg(config GlmCfg) {
	glmCfg = config
}

func GenerateToken(apikey string, expSeconds int) string {
	parts := strings.Split(apikey, ".")
	if len(parts) != 2 {
		logger.Error("=====apikey format error", "apikey", apikey)
		return ""
	}
	id := parts[0]
	secret := parts[1]
	now := time.Now().UnixMilli()
	payload := jwt.MapClaims{
		"api_key":   id,
		"timestamp": now,
		"exp":       now + int64(expSeconds)*1000,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		logger.Error("=====generate token error", "err", err)
		return ""
	}
	return signedToken
}

// joinSentences 分句
func joinSentences(s2 string, callback func(string)) {
	//logger.Info("=====model send back", "s", s2)
	if s2 == "  " {
		return
	}
	//logger.Info("====== cb", "s2", s2)
	ss := strings.Split(s2, "")

	for _, t := range ss {
		b := true
		for _, k := range Punctuations {
			if t == k {
				b = false
			}
		}
		if b {
			Results = append(Results, t)
		} else {
			Results = append(Results, t)
			fmt.Println(strings.Join(Results, ""))
			callback(strings.Join(Results, ""))
			Resultsn = append(Resultsn, strings.Join(Results, ""))
			Results = make([]string, 0)
		}
	}

}
