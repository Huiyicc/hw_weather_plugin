package main

/*
#include <stdio.h>
#include <stdlib.h>

inline void CallPluginLogFunc(void* ptr,const char*name,const char*raw) {
    typedef void (*Func)(const char*,const char*);
    if (ptr==NULL) {
        return;
    }
    Func func = (Func)ptr;
    func(name,raw);
};
inline void CallCallEinkFullUpdateImageFunc(void* ptr,unsigned char*data,int len) {
    typedef void (*Func)(unsigned char*data,int len);
    if (ptr==NULL) {
        return;
    }
    Func func = (Func)ptr;
    func(data,len);
};

*/
import "C"
import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"hw_weather_plugin/utils/utils"
	"hw_weather_plugin/weather"
	"io"
	"os"
	"strconv"
	"time"
	"unsafe"
)

// -----------------------------------

type PluginConfig struct {
	// 插件的目录
	Path  string `json:"plugin_path"`
	Calls []struct {
		// 调用的函数名
		Name string `json:"name"`
		// 调用的地址
		Address int64 `json:"address"`
	}
	CallsMap map[string]int64 `json:"-"`
}

type HWToolsPlugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

type ConfigUI struct {
	Widgets [][]Widget `json:"widgets"`
}

type Widget struct {
	Type   string  `json:"type"`
	Text   string  `json:"text"`
	Layout float64 `json:"layout"`
	Bind   string  `json:"bind"`
	Height float64 `json:"height"`
}

type ConfigPut struct {
	CityID             string `json:"city_id"`
	WeatherKey         string `json:"weather_key"`
	WeatherApiBusiness bool   `json:"weather_api_business"`
	EnableFahrenheit   bool   `json:"enable_fahrenheit"`
	AddiTitle          string `json:"addi_title"`
	AddiContent        string `json:"addi_content"`
}

type SubmitData struct {
	EventBind string    `json:"event_bind"`
	ConfigPut ConfigPut `json:"config"`
}

//-----------------------------------

const (
	PluginName = "生活插件"
)

var (
	pluginConfig  PluginConfig
	configPutData ConfigPut
	lastError     error

	//go:embed Description.txt
	description string
)

//-----------------------------------

const (

	// EinkFullUpdateImage 刷新墨水屏回调
	EinkFullUpdateImage = "EinkFullUpdateImage"
)

func CallPluginLogFunc(raw string) {
	pluginNameStr := C.CString(PluginName)
	defer C.free(unsafe.Pointer(pluginNameStr))
	rawStr := C.CString(raw)
	defer C.free(unsafe.Pointer(rawStr))
	C.CallPluginLogFunc(unsafe.Pointer(uintptr(pluginConfig.CallsMap["PluginLogInfo"])), pluginNameStr, rawStr)
}

func CallEinkFullUpdateImageFunc(data []byte) {
	cBytes := C.CBytes(data)
	defer C.free(cBytes)
	C.CallCallEinkFullUpdateImageFunc(unsafe.Pointer(uintptr(pluginConfig.CallsMap[EinkFullUpdateImage])), (*C.uchar)(cBytes), C.int(len(data)))
}

//-----------------------------------

//export PluginTest
func PluginTest() C.int {
	return C.int(0)
}

//export PluginRegister
func PluginRegister() *C.char {
	plugin := HWToolsPlugin{
		Name:        PluginName,
		Version:     "0.0.1",
		Author:      "回忆",
		Description: description,
	}
	data, _ := json.Marshal(plugin)
	return C.CString(string(data))
}

//export PluginUnRegister
func PluginUnRegister() bool {
	// 保存配置
	err := saveConfig([]byte{})
	if err != nil {
		lastError = err
	}
	return err == nil
}

//export PluginInit
func PluginInit(config *C.char) bool {
	err := json.Unmarshal([]byte(C.GoString(config)), &pluginConfig)
	if err != nil {
		lastError = err
		return false
	}
	pluginConfig.CallsMap = make(map[string]int64)
	for _, call := range pluginConfig.Calls {
		if call.Address == 0 {
			continue
		}
		pluginConfig.CallsMap[call.Name] = call.Address
	}
	f, err := os.Open(pluginConfig.Path + "plugin.config")
	if err != nil {
		f, err = os.Create(pluginConfig.Path + "plugin.config")
		if err != nil {
			lastError = err
			return false
		}
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	return json.Unmarshal(data, &configPutData) == nil
}

//export PluginConfigUI
func PluginConfigUI() *C.char {
	uis := ConfigUI{
		Widgets: [][]Widget{
			{
				{
					Type:   "text",
					Text:   "和风天气秘钥",
					Layout: 2,
				},
				{
					Type:   "input",
					Bind:   "weather_key",
					Text:   configPutData.WeatherKey,
					Layout: 7,
				},
			},
			{
				{
					Type:   "checkbox",
					Text:   "付费接口(免费订阅秘钥不能勾选)",
					Bind:   "weather_api_business",
					Layout: 5,
				},
			},
			{
				{
					Type:   "text",
					Text:   "如果秘钥为空则使用共享接口",
					Layout: 10,
				},
			},
			// ------------------------
			{
				{
					Type:   "divider",
					Text:   "",
					Layout: 10,
				},
			},
			{
				{
					Type:   "checkbox",
					Text:   "启用华氏度",
					Bind:   "enable_fahrenheit",
					Layout: 3,
				},
			},
			// ------------------------
			{
				{
					Type:   "divider",
					Text:   "",
					Layout: 10,
				},
			},
			{
				{
					Type:   "text",
					Text:   "城市ID",
					Layout: 2,
				},
				{
					Type:   "input",
					Bind:   "city_id",
					Text:   configPutData.CityID,
					Layout: 7,
				},
			},
			// ------------------------
			{
				{
					Type:   "divider",
					Text:   "",
					Layout: 10,
				},
			},
			{
				{
					Type:   "text",
					Text:   "标题",
					Layout: 2,
				},
				{
					Type:   "input",
					Text:   utils.Ifs(configPutData.AddiTitle == "", "待办", configPutData.AddiTitle),
					Bind:   "addi_title",
					Layout: 7,
				},
			},
			{
				{
					Type:   "text",
					Text:   "内容",
					Layout: 2,
				},
				{
					Type:   "input_ml",
					Bind:   "addi_content",
					Text:   utils.Ifs(configPutData.AddiContent == "", "暂无", configPutData.AddiContent),
					Height: 100,
					Layout: 7,
				},
			},
			// ------------------------
			{
				{
					Type:   "divider",
					Text:   "",
					Layout: 10,
				},
			},
			{
				{
					Type:   "submit",
					Bind:   "weather_update",
					Text:   "手动更新",
					Layout: 1,
				},
			},
		},
	}
	data, _ := json.Marshal(uis)
	return C.CString(string(data))
}

func saveConfig(data []byte) error {
	// 保存配置
	tmp := ConfigPut{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	configPutData = tmp
	os.Remove(pluginConfig.Path + "plugin.config")
	f, err := os.OpenFile(pluginConfig.Path+"plugin.config", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		f, err = os.Create(pluginConfig.Path + "plugin.config")
		if err != nil {
			return err
		}
	}
	defer f.Close()
	f.Write(data)
	return nil
}
func saveConfigStruct(data ConfigPut) error {
	// 保存配置
	configPutData = data
	os.Remove(pluginConfig.Path + "plugin.config")
	f, err := os.OpenFile(pluginConfig.Path+"plugin.config", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		f, err = os.Create(pluginConfig.Path + "plugin.config")
		if err != nil {
			return err
		}
	}
	defer f.Close()
	b, _ := json.Marshal(data)
	f.Write(b)
	return nil
}

//export PluginSaveConfig
func PluginSaveConfig(data *C.char) bool {
	// 保存配置
	config := C.GoString(data)
	//cfg := ConfigPut{}
	//json.Unmarshal([]byte(config), &cfg)
	// 保存配置
	err := saveConfig([]byte(config))
	if err != nil {
		lastError = err
	}
	return err == nil
}

func GetWeatherImage() ([]byte, error) {
	if configPutData.CityID == "" {
		err := errors.New("城市ID不能为空")
		lastError = err
		return nil, err
	}
	if configPutData.WeatherKey == "" {
		// 共享接口
		data, err := weather.DerawImage(
			configPutData.CityID,
			"",
			"",
			configPutData.AddiTitle,
			configPutData.AddiContent,
			configPutData.EnableFahrenheit,
		)
		if err != nil {
			lastError = err
			return nil, err
		}
		return data, nil
	} else {
		// 付费接口
		data, err := weather.DerawImage(
			configPutData.CityID,
			utils.Ifs(
				configPutData.WeatherApiBusiness,
				"https://api.qweather.com/v7",
				"https://devapi.qweather.com/v7",
			),
			configPutData.WeatherKey,
			configPutData.AddiTitle,
			configPutData.AddiContent,
			configPutData.EnableFahrenheit,
		)
		if err != nil {
			lastError = err
			return nil, err
		}
		return data, nil
	}

}

//export PluginTimedEvent
func PluginTimedEvent() bool {
	// CallPluginLogFunc(fmt.Sprintf("CheckUpdateStatus:%v", CheckUpdateStatus()))
	if !CheckUpdateStatus() {
		return true
	}
	data, err := GetWeatherImage()
	CallPluginLogFunc(fmt.Sprintf("插件定时事件"))
	if err != nil {
		lastError = err
		return false
	}
	CallEinkFullUpdateImageFunc(data)
	SetUpdateStatus(false)
	return true
}

//export PluginSubmit
func PluginSubmit(data *C.char) bool {
	rdata := C.GoString(data)
	var subData SubmitData
	json.Unmarshal([]byte(rdata), &subData)
	CallPluginLogFunc(fmt.Sprintf("插件提交事件:%s", string(rdata)))
	if subData.EventBind == "weather_update" {
		err := saveConfigStruct(subData.ConfigPut)
		if err != nil {
			lastError = err
			return false
		}
		imgData, err := GetWeatherImage()
		if err != nil {
			lastError = err
			return false
		}
		CallEinkFullUpdateImageFunc(imgData)
		return true
	}
	lastError = errors.New("未知事件")
	return false
}

//export PluginGetLastError
func PluginGetLastError() *C.char {
	if lastError == nil {
		return C.CString("")
	}
	auto := C.CString(lastError.Error())
	lastError = nil
	return auto
}

var updatesCache = make(map[string]bool)

// CheckUpdateStatus 检查更新状态
// 返回false不需要更新
func CheckUpdateStatus() bool {
	//CallPluginTestFunc(fmt.Sprintf("configPutData.CityID:%s", configPutData.CityID))
	if configPutData.CityID == "" {
		return false
	}
	nowTime := time.Now()
	nowHour := nowTime.Hour()
	nowDay := nowTime.Day()
	if nowHour < 4 {
		// 3点前
		nowHour = 0
	} else if nowHour < 9 {
		// 8点前
		nowHour = 8
	} else if nowHour < 14 {
		// 13点前
		nowHour = 13
	} else if nowHour < 20 {
		// 19点前
		nowHour = 19
	} else {
		// 19点后
		nowHour = 24
	}
	timeKey := "weather:" + configPutData.CityID + ":" + strconv.Itoa(nowDay) + ":" + strconv.Itoa(nowHour)
	c, ifSet := updatesCache[timeKey]
	if ifSet {
		return c
	}
	return true
}

func SetUpdateStatus(s bool) {
	if configPutData.CityID == "" {
		return
	}
	nowTime := time.Now()
	nowHour := nowTime.Hour()
	nowDay := nowTime.Day()
	if nowHour < 4 {
		// 3点前
		nowHour = 0
	} else if nowHour < 9 {
		// 8点前
		nowHour = 8
	} else if nowHour < 14 {
		// 13点前
		nowHour = 13
	} else if nowHour < 20 {
		// 19点前
		nowHour = 19
	} else {
		// 19点后
		nowHour = 24
	}
	timeKey := "weather:" + configPutData.CityID + ":" + strconv.Itoa(nowDay) + ":" + strconv.Itoa(nowHour)
	updatesCache = make(map[string]bool)
	updatesCache[timeKey] = s
}
