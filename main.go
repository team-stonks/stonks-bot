package handler

import (
	"encoding/json"
	"io"
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
const bytesInMegabyte int64 = 1048576

var chatId int = 0

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

type OuterMessage struct {
	Text string `json:"text"`
}

func parseTelegramRequest(body []byte) (*Update, error) {
	var update Update
	err := json.Unmarshal(body, &update)
	if err != nil {
		log.Printf("Couldn't decode incoming update. Error: %s", err.Error())
		return nil, err
	}
	if update.Message.Text == startCommand {
		chatId = update.Message.Chat.Id
	}
	return &update, nil
}

func parseOuterRequest(body []byte) (*OuterMessage, error) {
	var message OuterMessage
	err := json.Unmarshal(body, &message)
	if err != nil {
		log.Printf("Couldn't decode incoming outer HTTP request. Error: %s", err.Error())
		return nil, err
	}
	return &message, nil
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
		log.Printf("Error in parsing Telegram answer: %s", errRead.Error())
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Printf("Body of Telegram response: %s", bodyString)

	return bodyString, nil
}

func HandleOuterHTTPRequest(body []byte) {
	var outerMessage, errOuter = parseOuterRequest(body)

	if errOuter != nil {
		log.Printf("Couldn't parse outer HTTP request. Error: %s", errOuter.Error())
		return
	}

	var telegramResponseBody, errTelegram = sendTextToChat(chatId, string(body) /*outerMessage.Text*/)
	if errTelegram != nil {
		log.Printf("Got error %s from Telegram, response body is %s", errTelegram.Error(), telegramResponseBody)
	} else {
		log.Printf("Message [%s] successfuly distributed to chat id [%d]", outerMessage.Text, chatId)
	}
}

func HandleTelegramWebHook(w http.ResponseWriter, r *http.Request) {
	requestBodyBytes, errRead := ioutil.ReadAll(io.LimitReader(r.Body, bytesInMegabyte))
	if errRead != nil {
		log.Printf("Couldn't read request body. Error: %s", errRead.Error())
		return
	}

	var update, errUpdate = parseTelegramRequest(requestBodyBytes)

	if errUpdate != nil {
		log.Printf("Couldn't parse Telegram update. Error: %s", errUpdate.Error())
		return
	}

	if update.Message.Chat.Id == 0 {
		HandleOuterHTTPRequest(requestBodyBytes)
		return
	}

	message := "Hello world!"

	var telegramResponseBody, errTelegram = sendTextToChat(chatId, message)
	if errTelegram != nil {
		log.Printf("Got error %s from Telegram, response body is %s", errTelegram.Error(), telegramResponseBody)
	} else {
		log.Printf("Message [%s] successfuly distributed to chat id [%d]", message, chatId)
	}
}
