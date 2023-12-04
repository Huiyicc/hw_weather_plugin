package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type weatherResp struct {
	Error int         `json:"error"`
	Data  WeatherResp `json:"data"`
}

type WeatherResp struct {
	UpdateTime    string        `json:"updateTime"`
	WeatherStatus WeatherStatus `json:"weather_status"`
	WeatherIndexs WeatherIndexs `json:"weather_indexs"`
}

func (c *WeatherResp) Parse(status *CityWeatherInfo, indexs *CityWeatherIndexInfo) {
	c.WeatherStatus = status.Now
	c.WeatherIndexs = indexs.Index
	c.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
}

type WeatherStatus struct {
	ObsTime   string `json:"obsTime"`   // 数据观测时间
	Temp      string `json:"temp"`      // 温度，默认单位：摄氏度
	FeelsLike string `json:"feelsLike"` // 体感温度，默认单位：摄氏度
	Icon      string `json:"icon"`      // 天气状况图标代码
	Text      string `json:"text"`      // 天气状况的文字描述，包括阴晴雨雪等天气状态的描述
	Wind360   string `json:"wind360"`   // 风向360角度
	WindDir   string `json:"windDir"`   // 风向
	WindScale string `json:"windScale"` // 风力等级
	WindSpeed string `json:"windSpeed"` // 风速，公里/小时
	Humidity  string `json:"humidity"`  // 相对湿度，百分比数值
	Precip    string `json:"precip"`    // 当前小时累计降水量，默认单位：毫米
	Pressure  string `json:"pressure"`  // 大气压强，默认单位：百帕
	Vis       string `json:"vis"`       // 能见度，默认单位：公里
	Cloud     string `json:"cloud"`     // 云量，百分比数值
	Dew       string `json:"dew"`       // 露点温度.可能为空
}
type WeatherIndexs struct {
	Motion    WeatherIndexStatus `json:"motion"`    // 运动指数
	CarWash   WeatherIndexStatus `json:"carWash"`   // 洗车指数
	Dress     WeatherIndexStatus `json:"dress"`     // 穿衣指数
	Fishing   WeatherIndexStatus `json:"fishing"`   // 钓鱼指数
	UV        WeatherIndexStatus `json:"uv"`        // 紫外线指数
	Travel    WeatherIndexStatus `json:"travel"`    // 旅游指数
	Allergy   WeatherIndexStatus `json:"allergy"`   // 过敏指数
	Comfort   WeatherIndexStatus `json:"comfort"`   // 舒适度指数
	Cold      WeatherIndexStatus `json:"cold"`      // 感冒指数
	Air       WeatherIndexStatus `json:"air"`       // 空气污染扩散条件指数
	Ac        WeatherIndexStatus `json:"ac"`        // 空调开启指数
	Sunglass  WeatherIndexStatus `json:"sunglass"`  // 太阳镜指数
	Makeup    WeatherIndexStatus `json:"makeup"`    // 化妆指数
	Dry       WeatherIndexStatus `json:"dry"`       // 晾晒指数
	Traffic   WeatherIndexStatus `json:"traffic"`   // 交通指数
	Sunscreen WeatherIndexStatus `json:"sunscreen"` // 防晒指数
}
type WeatherIndexStatus struct {
	Name     string `json:"name"`
	Level    string `json:"level"`
	Category string `json:"category"`
	Text     string `json:"text"`
}

// GetWeather 获取天气信息
func GetWeather(cityID string) (WeatherResp, error) {
	ret := weatherResp{}
	resp, err := http.Post("http://openapi.hyiy.top/api/life/weather", "application/x-www-form-urlencoded", strings.NewReader("cityID="+cityID))
	if err != nil {
		return ret.Data, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret.Data, err
	}
	err = json.Unmarshal(respData, &ret)
	return ret.Data, err
}
