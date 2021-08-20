package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type wallet map[string]float64
var database = map[int]wallet{}
type bResponse struct {
	Symbol string `json:"symbol"`
	Price float64 `json:"price,string"`
}

func GetExchangeRate(currency string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", currency)
	answ, err := http.Get(url)
	if err != nil {
		// URL NOT FOUND
		return 7, err
	}

	var bRes bResponse
	err = json.NewDecoder(answ.Body).Decode(&bRes)
	if err != nil {
		return 8, err
	}

	if bRes.Symbol == "" {
		return 9, errors.New("Невалидная валюта")
	}

	// RUB
	url = fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=USDTRUB")
	answ, err = http.Get(url)
	if err != nil {
		// URL NOT FOUND
		return 7, err
	}

	var dollar_rate float64 = bRes.Price
	err = json.NewDecoder(answ.Body).Decode(&bRes)
	if err != nil {
		return 8, err
	}

	return bRes.Price * dollar_rate, nil
}

func main() {
	bot, err := tgbotapi.NewBotAPI("1959121735:AAGRTHLoSryFmi_okNee3PbyVeDii51R3a0")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	// log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		message_str := update.Message.Text
		// username_str := update.Message.From.UserName
		userID := update.Message.From.ID

		command := strings.Split(message_str, " ")
		switch command[0] {
		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы"))
				continue
			}

			_, err := GetExchangeRate(command[1])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Невалидная валюта"))
				continue
			}

			val, err := strconv.ParseFloat(command[2], 64)
			if err == nil {
				if _, ok := database[userID]; !ok {
					database[userID] = make(wallet)
					database[userID][command[1]] = val
				} else {
					database[userID][command[1]] += val
				}

				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Баланс изменен"))
			} else {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная ошибка"))
			}
		case "SUB":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы"))
				continue
			}

			_, err := GetExchangeRate(command[1])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Невалидная валюта"))
				continue
			}

			val, err := strconv.ParseFloat(command[2], 64)
			if err == nil {
				//  or _, ok_ := database[userID][command[1]]; !ok_
				if _, ok := database[userID]; !ok {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
						"Невозможно изменить баланс: баланс не инициализирован"))
				} else {
					database[userID][command[1]] -= math.Min(val, database[userID][command[1]])
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Баланс изменен"))
				}
			} else {
				// bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная ошибка"))
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы"))
				continue
			}
			delete(database[userID], command[1])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("Удален баланс для %s\n", command[1])))
		case "SHOW":
			output := "Ваш баланс:"
			if len(database[userID]) != 0 {
				output += "\n"
				for key, value := range database[userID] {
					ExPrice, err := GetExchangeRate(key)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
						output += fmt.Sprintf("%s: %.3f\n", key, value)
						continue
					}
					output += fmt.Sprintf("%s: %.3f, %.4f RUB\n", key, value, ExPrice * value)
				}
			} else {
				output += " не инициализирован"
			}

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, output))
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
		}
	}
}