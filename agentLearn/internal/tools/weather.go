package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// WeatherTool 天气查询工具（使用 OpenMeteo 免费 API）
type WeatherTool struct {
	client *resty.Client
}

// NewWeatherTool 创建天气查询工具
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{
		client: resty.New().SetBaseURL("https://api.open-meteo.com"),
	}
}

// Name 工具名称
func (t *WeatherTool) Name() string {
	return "weather_query"
}

// Description 工具描述
func (t *WeatherTool) Description() string {
	return "查询全球任意地点的实时天气和天气预报。提供温度、湿度、风速、降水等信息。无需 API Key，完全免费。"
}

// Parameters 工具参数定义
func (t *WeatherTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"city": map[string]interface{}{
				"type":        "string",
				"description": "城市名称（例如：Beijing, Shanghai, New York）",
			},
			"country": map[string]interface{}{
				"type":        "string",
				"description": "国家代码（可选，例如：CN, US, JP）",
			},
			"days": map[string]interface{}{
				"type":        "integer",
				"description": "预报天数（1-7 天，默认为 1 天）",
				"minimum":     1,
				"maximum":     7,
			},
		},
		"required": []string{"city"},
	}
}

// Execute 执行天气查询
func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	city, ok := args["city"].(string)
	if !ok || city == "" {
		return "", fmt.Errorf("city is required")
	}

	country, _ := args["country"].(string)
	days := 1
	if d, ok := args["days"].(float64); ok {
		days = int(d)
		if days < 1 {
			days = 1
		}
		if days > 7 {
			days = 7
		}
	}

	// 第一步：通过地理编码获取经纬度
	geocodingURL := "https://geocoding-api.open-meteo.com/v1/search"
	var geoResp struct {
		Results []struct {
			Name      string  `json:"name"`
			Country   string  `json:"country"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"results"`
	}

	geoReq := t.client.R().SetContext(ctx)
	if country != "" {
		geoReq.SetQueryParam("country", country)
	}
	geoReq.SetQueryParams(map[string]string{
		"name":     city,
		"count":    "1",
		"language": "zh",
		"format":   "json",
	})

	geoRespBody, err := geoReq.Get(geocodingURL)
	if err != nil {
		return "", fmt.Errorf("地理编码查询失败：%w", err)
	}

	if err := json.Unmarshal(geoRespBody.Body(), &geoResp); err != nil {
		return "", fmt.Errorf("解析地理编码响应失败：%w", err)
	}

	if len(geoResp.Results) == 0 {
		return "", fmt.Errorf("未找到城市：%s", city)
	}

	location := geoResp.Results[0]
	lat := location.Latitude
	lon := location.Longitude
	cityName := location.Name
	countryName := location.Country

	// 第二步：查询天气数据
	weatherURL := "https://api.open-meteo.com/v1/forecast"
	var weatherResp struct {
		Daily struct {
			Time             []string  `json:"time"`
			TemperatureMax   []float64 `json:"temperature_2m_max"`
			TemperatureMin   []float64 `json:"temperature_2m_min"`
			PrecipitationSum []float64 `json:"precipitation_sum"`
			WeatherCode      []int     `json:"weathercode"`
		} `json:"daily"`
		Current struct {
			Time             string  `json:"time"`
			Temperature      float64 `json:"temperature_2m"`
			RelativeHumidity int     `json:"relative_humidity_2m"`
			WindSpeed        float64 `json:"wind_speed_10m"`
			WindDirection    float64 `json:"wind_direction_10m"`
			WeatherCode      int     `json:"weathercode"`
		} `json:"current"`
	}

	weatherReq := t.client.R().SetContext(ctx).SetQueryParams(map[string]string{
		"latitude":           fmt.Sprintf("%f", lat),
		"longitude":          fmt.Sprintf("%f", lon),
		"current":            "temperature_2m,relative_humidity_2m,wind_speed_10m,wind_direction_10m,weathercode",
		"daily":              "temperature_2m_max,temperature_2m_min,precipitation_sum,weathercode",
		"temperature_unit":   "celsius",
		"wind_speed_unit":    "kmh",
		"precipitation_unit": "mm",
		"timezone":           "auto",
		"forecast_days":      fmt.Sprintf("%d", days),
	})

	weatherRespBody, err := weatherReq.Get(weatherURL)
	if err != nil {
		return "", fmt.Errorf("天气查询失败：%w", err)
	}

	if err := json.Unmarshal(weatherRespBody.Body(), &weatherResp); err != nil {
		return "", fmt.Errorf("解析天气响应失败：%w", err)
	}

	// 第三步：格式化输出结果
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📍 地点：%s, %s (纬度：%.2f, 经度：%.2f)\n\n", cityName, countryName, lat, lon))

	// 当前天气
	sb.WriteString("🌡️ 当前天气：\n")
	sb.WriteString(fmt.Sprintf("   温度：%.1f°C\n", weatherResp.Current.Temperature))
	sb.WriteString(fmt.Sprintf("   湿度：%d%%\n", weatherResp.Current.RelativeHumidity))
	sb.WriteString(fmt.Sprintf("   风速：%.1f km/h", weatherResp.Current.WindSpeed))
	if weatherResp.Current.WindDirection >= 0 {
		direction := getWindDirection(weatherResp.Current.WindDirection)
		sb.WriteString(fmt.Sprintf(" (%s)", direction))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("   天气：%s\n\n", getWeatherDescription(weatherResp.Current.WeatherCode)))

	// 天气预报
	if days > 1 && len(weatherResp.Daily.Time) > 0 {
		sb.WriteString("📅 天气预报：\n")
		for i := 0; i < len(weatherResp.Daily.Time) && i < days; i++ {
			date, _ := time.Parse("2006-01-02", weatherResp.Daily.Time[i])
			dateStr := date.Format("01 月 02 日")

			maxTemp := weatherResp.Daily.TemperatureMax[i]
			minTemp := weatherResp.Daily.TemperatureMin[i]
			precip := weatherResp.Daily.PrecipitationSum[i]
			weatherCode := weatherResp.Daily.WeatherCode[i]

			sb.WriteString(fmt.Sprintf("   %s: %s  %.0f°C ~ %.0f°C",
				dateStr,
				getWeatherDescription(weatherCode),
				minTemp,
				maxTemp))

			if precip > 0 {
				sb.WriteString(fmt.Sprintf("  🌧️ 降水：%.1fmm", precip))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n💡 数据来源：Open-Meteo (免费天气 API)")

	return sb.String(), nil
}

// getWeatherDescription 根据天气代码返回描述
func getWeatherDescription(code int) string {
	weatherCodes := map[int]string{
		0:  "☀️ 晴朗",
		1:  "🌤️ 主要晴朗",
		2:  "⛅ 部分多云",
		3:  "☁️ 阴天",
		45: "🌫️ 雾",
		48: "🌫️ 雾凇",
		51: "🌦️ 轻度毛毛雨",
		53: "🌦️ 中度毛毛雨",
		55: "🌧️ 重度毛毛雨",
		61: "🌦️ 轻度雨",
		63: "🌧️ 中雨",
		65: "🌧️ 大雨",
		71: "🌨️ 轻度雪",
		73: "🌨️ 中雪",
		75: "❄️ 大雪",
		77: "🌨️ 雪粒",
		80: "🌦️ 轻度阵雨",
		81: "🌧️ 中雨",
		82: "⛈️ 暴雨",
		85: "🌨️ 轻度阵雪",
		86: "🌨️ 重度阵雪",
		95: "⛈️ 雷暴",
		96: "⛈️ 雷暴伴轻度冰雹",
		99: "⛈️ 雷暴伴重度冰雹",
	}

	if desc, exists := weatherCodes[code]; exists {
		return desc
	}
	return "🌡️ 未知天气"
}

// getWindDirection 根据风向角度返回方向
func getWindDirection(degrees float64) string {
	directions := []string{
		"北", "东北", "东", "东南", "南", "西南", "西", "西北",
	}
	index := int(math.Mod(degrees+22.5, 360) / 45)
	return directions[index]
}
