package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type WeatherResponse struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Rain struct {
		OneHour float64 `json:"1h"`
	} `json:"rain"`
}

func main() {
	// Загружаем переменные окружения
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	// Инициализируем бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Бот авторизован как %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)

	updates := bot.GetUpdatesChan(updateConfig)

	// Создаем карту для отслеживания состояния пользователей
	userStates := make(map[int64]string)

	// Обработка сообщений
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		chatID := update.Message.Chat.ID

		command := update.Message.Command()
		text := update.Message.Text

		switch {
		case command == "start":
			msg.Text = "Привет! Я бот погоды.\nИспользуйте команду /weather для получения погоды."

		case command == "weather":
			userStates[chatID] = "awaiting_city"
			msg.Text = "Укажите город."

		case command == "help":
			msg.Text = "Доступные команды:\n" +
				"/weather - узнать погоду\n" +
				"/help - показать это сообщение"

		case userStates[chatID] == "awaiting_city":
			weather := getWeather(text)
			msg.Text = weather
			delete(userStates, chatID)

		default:
			msg.Text = "Используйте команду /weather для получения погоды"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
	}
}

func getWeather(city string) string {
	weatherApiKey := os.Getenv("WEATHER_API_KEY")
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric&lang=ru", city, weatherApiKey)

	resp, err := http.Get(url)
	if err != nil {
		return "Ошибка при получении данных о погоде"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "Город не найден. Проверьте правильность написания."
	}

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return "Ошибка при обработке данных о погоде"
	}

	rainInfo := "нет"
	if weather.Rain.OneHour > 0 {
		rainInfo = fmt.Sprintf("%.1f мм/ч", weather.Rain.OneHour)
	}

	return fmt.Sprintf("Погода в %s:\n"+
		"🌡 Температура: %.1f°C\n"+
		"💨 Ветер: %.1f м/с\n"+
		"💧 Влажность: %d%%\n"+
		"🌧 Осадки: %s\n"+
		"☁️ Условия: %s",
		city,
		weather.Main.Temp,
		weather.Wind.Speed,
		weather.Main.Humidity,
		rainInfo,
		weather.Weather[0].Description)
}
