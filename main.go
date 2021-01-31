package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const startCommand string = "/start"

const telegramApiBaseUrl string = "https://api.telegram.org/bot"
const telegramApiSendMessage string = "/sendMessage"
const telegramTokenEnv string = "TELEGRAM_BOT_TOKEN"

var telegramApi string = telegramApiBaseUrl + os.Getenv(telegramTokenEnv) + telegramApiSendMessage

type Chat struct {
	Id int `json:"id"`
}

type Message struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
}

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

func parseRequest(r *http.Request) (*Update, error) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Couldn't decode incoming update %s", err.Error())
		return nil, err
	}
	return &update, nil
}

func sendTextToChat(chatId int, text string) (string, error) {
	log.Printf("Sending message [%s] to chat_id [%d]", text, chatId)
	response, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"text":    {text},
		})

	if err != nil {
		log.Printf("Error when posting text to the chat: %s", err.Error())
		return "", err
	}
	defer response.Body.Close()

	var bodyBytes, errRead = ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("Error in parsing Telegram answer %s", errRead.Error())
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Printf("Body of Telegram response: %s", bodyString)

	return bodyString, nil
}

func HandleTelegramWebHook(w http.ResponseWriter, r *http.Request) {

	var update, err = parseRequest(r)
	if err != nil {
		log.Printf("Error parsing update, %s", err.Error())
		return
	}

	message := "Hello world!"

	var telegramResponseBody, errTelegram = sendTextToChat(update.Message.Chat.Id, message)
	if errTelegram != nil {
		log.Printf("Got error %s from Telegram, response body is %s", errTelegram.Error(), telegramResponseBody)
	} else {
		log.Printf("Message [%s] successfuly distributed to chat id [%d]", message, update.Message.Chat.Id)
	}
}
