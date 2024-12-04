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
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
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
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	// Обработка сообщений
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch update.Message.Text {
		case "/weather":
			weather := getWeather()
			msg.Text = weather
		case "/photo":
			resp, err := http.Get("https://www.zoo-mega.ru/_mod_files/ce_images/news/107-min.jpg")
			if err != nil {
				log.Println(err)
				msg.Text = "Не могу получить фото :("
				break
			}
			func() {
				defer resp.Body.Close()
				photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileReader{
					Name:   "photo.jpg",
					Reader: resp.Body,
				})
				_, err = bot.Send(photo)
			}()
			if err != nil {
				log.Println(err)
				msg.Text = "Не могу отправить фото :("
			}
		default:
			msg.Text = "Используйте /weather для получения информации о погоде"
		}

		bot.Send(msg)
	}
}

func getWeather() string {
	weatherApiKey := os.Getenv("WEATHER_API_KEY")
	// Здесь используется Москва как пример (можно изменить город)
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=Moscow&appid=%s&units=metric", weatherApiKey)

	resp, err := http.Get(url)
	if err != nil {
		return "Ошибка при получении данных о погоде"
	}
	defer resp.Body.Close()

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return "Ошибка при обработке данных о погоде"
	}

	return fmt.Sprintf("Температура: %.1f°C\nПогода: %s",
		weather.Main.Temp,
		weather.Weather[0].Description)
}
