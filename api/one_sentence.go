package api

import (
	_ "embed"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
)

type OneSentenceData struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid"`
	Hitokoto   string `json:"hitokoto"`
	Type       string `json:"type"`
	From       string `json:"from"`
	FromWho    string `json:"from_who"`
	Creator    string `json:"creator"`
	CreatorUid int    `json:"creator_uid"`
	Reviewer   int    `json:"reviewer"`
	CommitFrom string `json:"commit_from"`
	CreatedAt  string `json:"created_at"`
	Length     int    `json:"length"`
}

var (
	//go:embed hitokoto.json
	hitokotoRaw []byte
)

// GetOneSentence 用于获取一言
func GetOneSentence() (OneSentenceData, error) {
	ret := OneSentenceData{}
	req, err := http.Get("https://v1.hitokoto.cn/?c=i&max_length=16&min_length=15")
	if err != nil {
		return ret, err
	}
	defer req.Body.Close()
	respData, err := io.ReadAll(req.Body)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(respData, &ret)
	return ret, err
}

func GetOneSentenceLocal() (OneSentenceData, error) {
	var hitokotos []OneSentenceData

	// 解析 JSON 数据到 hitokotos 切片
	err := json.Unmarshal(hitokotoRaw, &hitokotos)
	if err != nil {
		return OneSentenceData{}, err
	}

	// 生成一个随机索引
	index := rand.Intn(len(hitokotos))

	// 返回随机选中的 OneSentenceData 实例
	return hitokotos[index], nil
}
