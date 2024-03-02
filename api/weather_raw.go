package api

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// 城市数据
type cityInteface struct {
	citys     []citys
	Datas     map[string]cityCache
	DatasList map[string]cityCache // 单层字典,键为城市代码/值为地区
}
type citys struct {
	Iso3166   string   `json:"ISO_3166"`
	CountryEN string   `json:"Country_EN"`
	CountryCN string   `json:"Country_CN"`
	Regions   []Region `json:"Regions"`
}
type Region struct {
	Name   string `json:"Name"`
	NameEn string `json:"Name_EN"`
	Citys  []City `json:"Citys"`
}
type City struct {
	Name      string     `json:"Name"`
	NameEn    string     `json:"Name_EN"`
	Locations []Location `json:"Locations"`
}

type Location struct {
	LocationID string `json:"LocationID"`
	LocationEN string `json:"Location_EN"`
	Location   string `json:"Location"`
	Latitude   string `json:"Latitude"`
	Longitude  string `json:"Longitude"`
	Adcode     string `json:"Adcode"`
}

type cityCache struct {
	Name string
	Code string
	Son  map[string]cityCache // 一层层嵌套的字典,键为省/城市/地区名
}

type CityWeatherInfo struct {
	Code       string        `json:"code"`
	UpdateTime string        `json:"updateTime"`
	FxLink     string        `json:"fxLink"`
	Now        WeatherStatus `json:"now"`
}

var (
	//go:embed city_list.json
	cityRaw []byte

	cityDatas cityInteface
)

func initWeatherData() error {
	err := json.Unmarshal(cityRaw, &cityDatas.citys)
	if err != nil {
		return err
	}
	cityDatas.Datas = make(map[string]cityCache)
	cityDatas.DatasList = make(map[string]cityCache)
	for _, cityData := range cityDatas.citys {
		for _, province := range cityData.Regions {
			if _, ifSet := cityDatas.Datas[province.Name]; !ifSet {
				cityDatas.Datas[province.Name] = cityCache{
					Name: province.Name,
					Code: "",
					Son:  make(map[string]cityCache),
				}
			}
			for _, city := range province.Citys {
				if _, ifSet := cityDatas.Datas[province.Name].Son[city.Name]; !ifSet {
					cityDatas.Datas[province.Name].Son[city.Name] = cityCache{
						Name: city.Name,
						Code: "",
						Son:  make(map[string]cityCache),
					}
				}
				for _, county := range city.Locations {
					if _, ifSet := cityDatas.Datas[province.Name].Son[city.Name].Son[county.Location]; !ifSet {
						cityDatas.Datas[province.Name].Son[city.Name].Son[county.Location] = cityCache{
							Name: county.Location,
							Code: county.LocationID,
							Son:  nil,
						}
						cityDatas.DatasList[county.LocationID] = cityDatas.Datas[province.Name].Son[city.Name].Son[county.Location]
					}
				}
			}
		}
	}
	return nil
}

func checkRespCode(code string) error {
	if code == "200" {
		return nil
	} else if code == "204" {
		return errors.New("城市数据不存在")
	} else if code == "400" {
		return errors.New("请求错误")
	} else if code == "401" {
		return errors.New("认证失败,请联系管理员")
	} else if code == "402" {
		return errors.New("超过访问次数,请联系管理员")
	} else if code == "403" {
		return errors.New("无访问权限,请联系管理员")
	} else if code == "404" {
		return errors.New("数据或地区不存在")
	} else if code == "429" {
		return errors.New("超过限制访问次数,请稍后再试")
	} else if code == "500" {
		return errors.New("服务器内部错误,请联系管理员")
	} else {
		return errors.New(fmt.Sprintf("未知错误,错误码:%s", code))
	}
}

// GetCurrentWeather 获取当前天气
func GetCurrentWeather(cityID, host, key string) (ret CityWeatherInfo, raw []byte, err error) {
	if len(cityDatas.citys) == 0 {
		if err := initWeatherData(); err != nil {
			return ret, nil, err
		}
	}
	if _, ifSet := cityDatas.DatasList[cityID]; !ifSet && len(key) == 0 {
		return ret, nil, errors.New("城市ID不存在\n请注意,国际城市ID需要使用开发或付费接口")
	}
	url := fmt.Sprintf("%s/weather/now?location=%s&key=%s",
		host,
		cityID,
		key,
	)
	resp, err := http.Get(url)
	if err != nil {
		return ret, nil, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret, nil, err
	}
	err = json.Unmarshal(respData, &ret)
	if err != nil {
		return ret, nil, err
	}
	err = checkRespCode(ret.Code)
	if err != nil {
		return ret, nil, err
	}
	return ret, respData, nil
}

type CityWeatherIndexInfo struct {
	Code       string        `json:"code"`
	UpdateTime string        `json:"updateTime"`
	FxLink     string        `json:"fxLink"`
	Index      WeatherIndexs `json:"index"` // 生活指数
}

func convertWeatherIndex(data []byte) (CityWeatherIndexInfo, error) {
	type cityWeatherIndexRaw struct {
		Code       string `json:"code"`
		UpdateTime string `json:"updateTime"`
		FxLink     string `json:"fxLink"`
		Daily      []struct {
			Date     string `json:"date"`
			Type     string `json:"type"`
			Name     string `json:"name"`
			Level    string `json:"level"`
			Category string `json:"category"`
			Text     string `json:"text"`
		} `json:"daily"`
	}
	ret := CityWeatherIndexInfo{}
	raw := cityWeatherIndexRaw{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return ret, err
	}
	err = checkRespCode(raw.Code)
	if err != nil {
		return ret, err
	}
	ret.Code = raw.Code
	ret.UpdateTime = raw.UpdateTime
	ret.FxLink = raw.FxLink
	for _, v := range raw.Daily {
		var tmpStatus *WeatherIndexStatus
		switch v.Type {
		case "1":
			// 运动指数
			tmpStatus = &ret.Index.Motion
		case "2":
			// 洗车指数
			tmpStatus = &ret.Index.CarWash
		case "3":
			// 穿衣指数
			tmpStatus = &ret.Index.Dress
		case "4":
			// 钓鱼指数
			tmpStatus = &ret.Index.Fishing
		case "5":
			// 紫外线指数
			tmpStatus = &ret.Index.UV
		case "6":
			// 旅游指数
			tmpStatus = &ret.Index.Travel
		case "7":
			// 过敏指数
			tmpStatus = &ret.Index.Allergy
		case "8":
			// 舒适度指数
			tmpStatus = &ret.Index.Comfort
		case "9":
			// 感冒指数
			tmpStatus = &ret.Index.Cold
		case "10":
			// 空气污染扩散条件指数
			tmpStatus = &ret.Index.Air
		case "11":
			// 空调开启指数
			tmpStatus = &ret.Index.Ac
		case "12":
			// 太阳镜指数
			tmpStatus = &ret.Index.Sunglass
		case "13":
			// 化妆指数
			tmpStatus = &ret.Index.Makeup
		case "14":
			// 晾晒指数
			tmpStatus = &ret.Index.Dry
		case "15":
			// 交通指数
			tmpStatus = &ret.Index.Traffic
		case "16":
			// 防晒指数
			tmpStatus = &ret.Index.Sunscreen
		default:
			continue
		}
		tmpStatus.Name = v.Name
		tmpStatus.Level = v.Level
		tmpStatus.Category = v.Category
		tmpStatus.Text = v.Text
	}
	return ret, nil
}

// GetWeatherIndex 获取当前天气
//
// cityID: 城市ID
// 内置限流,每秒10个令牌,令牌桶容量30个
func GetWeatherIndex(cityID, host, key string) (ret CityWeatherIndexInfo, raw []byte, err error) {
	if len(cityDatas.citys) == 0 {
		if err := initWeatherData(); err != nil {
			return ret, nil, err
		}
	}
	if _, ifSet := cityDatas.DatasList[cityID]; !ifSet && len(key) == 0 {
		return ret, nil, errors.New("城市ID不存在\n请注意,国际城市ID需要使用开发或付费接口")
	}
	url := fmt.Sprintf("%s/indices/1d?type=0&location=%s&key=%s",
		host,
		cityID,
		key,
	)
	resp, err := http.Get(url)
	if err != nil {
		return ret, nil, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret, nil, err
	}
	ret, err = convertWeatherIndex(respData)
	if err != nil {
		return ret, nil, err
	}
	return ret, respData, nil
}
