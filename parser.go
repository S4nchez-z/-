package parser

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"weather-tracker/internal/types"

	"github.com/PuerkitoBio/goquery"
)

const (
	userAgent        = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	yandexWeatherURL = "https://yandex.ru/pogoda/ru/moscow"
)

type Parser struct {
	client *http.Client
}

func NewParser() *Parser {
	return &Parser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *Parser) GetWeather() (*types.Weather, error) {
	req, err := http.NewRequest("GET", yandexWeatherURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неверный статус код: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга HTML: %w", err)
	}

	return p.parseWeatherData(doc)
}

func (p *Parser) parseWeatherData(doc *goquery.Document) (*types.Weather, error) {
	weather := &types.Weather{}

	doc.Find(".temp__value").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			weather.Temperature = strings.TrimSpace(s.Text())
		}
	})

	doc.Find(".link__condition").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			weather.Condition = strings.TrimSpace(s.Text())
		}
	})

	doc.Find(".term__value").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if strings.Contains(text, "°") && weather.FeelsLike == "" {
			weather.FeelsLike = text
		}
	})

	doc.Find(".term[data-term='pressure'] .term__value").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			weather.Pressure = strings.TrimSpace(s.Text())
		}
	})

	doc.Find(".term[data-term='humidity'] .term__value").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			weather.Humidity = strings.TrimSpace(s.Text())
		}
	})

	doc.Find(".wind-speed").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			weather.Wind = strings.TrimSpace(s.Text())
		}
	})

	weather.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	if weather.Temperature == "" {
		return nil, fmt.Errorf("не удалось получить данные о погоде")
	}

	return weather, nil
}

func (p *Parser) PrintWeather(weather *types.Weather) {
	fmt.Println("┌─────────────────────────────────────────┐")
	fmt.Println("│           Погода в Москве               │")
	fmt.Println("├─────────────────────────────────────────┤")
	fmt.Printf("│  Температура: %-25s │\n", weather.Temperature)
	fmt.Printf("│  Состояние: %-27s │\n", weather.Condition)

	if weather.FeelsLike != "" {
		fmt.Printf("│  Ощущается как: %-22s │\n", weather.FeelsLike)
	}
	if weather.Pressure != "" {
		fmt.Printf("│  Давление: %-26s │\n", weather.Pressure)
	}
	if weather.Humidity != "" {
		fmt.Printf("│  Влажность: %-25s │\n", weather.Humidity)
	}
	if weather.Wind != "" {
		fmt.Printf("│  Ветер: %-29s │\n", weather.Wind)
	}

	fmt.Printf("│  Обновлено: %-26s │\n", weather.UpdatedAt)
	fmt.Println("└─────────────────────────────────────────┘")
}
