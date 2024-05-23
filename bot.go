package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

// pushMsg: Push message to LINE server.
func pushMsg(target, text string) error {
	if _, err := bot.PushMessage(
		&messaging_api.PushMessageRequest{
			To: target,
			Messages: []messaging_api.MessageInterface{
				&messaging_api.TextMessage{
					Text: text,
				},
			},
		},
		"",
	); err != nil {
		return err
	}
	return nil
}

// replyText: Reply text message to LINE server.
func replyText(replyToken, text string) error {
	if _, err := bot.ReplyMessage(
		&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				&messaging_api.TextMessage{
					Text: text,
				},
			},
		},
	); err != nil {
		return err
	}
	return nil
}

// callbackHandler: Handle callback from LINE server.
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	cb, err := webhook.ParseRequest(os.Getenv("ChannelSecret"), r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range cb.Events {
		log.Printf("Got event %v", event)
		switch e := event.(type) {
		case webhook.MessageEvent:
			// 取得用戶 ID
			var uID string
			switch source := e.Source.(type) {
			case webhook.UserSource:
				uID = source.UserId
			case webhook.GroupSource:
				uID = source.UserId
			case webhook.RoomSource:
				uID = source.UserId
			}
			log.Println("User ID:", uID)
			fireDB.SetPath(fmt.Sprintf("%s/%s", DBExpensePath, uID))

			switch message := e.Message.(type) {
			// Handle only on text message
			case webhook.TextMessageContent:
				// Pass message text to Gemini API for FunctionCall.
				response := gemini.GeminiFunctionCall(message.Text)
				replyText(e.ReplyToken, response)

			// Handle only on Sticker message
			case webhook.StickerMessageContent:
				var kw string
				for _, k := range message.Keywords {
					kw = kw + "," + k
				}

				outStickerResult := fmt.Sprintf("收到貼圖訊息: %s, pkg: %s kw: %s  text: %s", message.StickerId, message.PackageId, kw, message.Text)
				if err := replyText(e.ReplyToken, outStickerResult); err != nil {
					log.Print(err)
				}

			// Handle only image message
			case webhook.ImageMessageContent:
				log.Println("Got img msg ID:", message.Id)

				//Get image binary from LINE server based on message ID.
				// data, err := GetImageBinary(blob, message.Id)
				// if err != nil {
				// 	log.Println("Got GetMessageContent err:", err)
				// 	continue
				// }

				// ret, err := gemini.GeminiImage(data, ImagePrompt)
				// if err != nil {
				// 	ret = "無法辨識影片內容文字，請重新輸入:" + err.Error()
				// }

			// Handle only video message
			case webhook.VideoMessageContent:
				log.Println("Got video msg ID:", message.Id)

			default:
				log.Printf("Unknown message: %v", message)
			}
		case webhook.PostbackEvent:
			// Using urls value to parse event.Postback.Data strings.
			ret, err := url.ParseQuery(e.Postback.Data)
			if err != nil {
				log.Print("action parse err:", err, " dat=", e.Postback.Data)
				continue
			}
			log.Println("Action:", ret["action"])
			log.Println("Calc calories m_id:", ret["m_id"])

		case webhook.FollowEvent:
			log.Printf("message: Got followed event")
		case webhook.BeaconEvent:
			log.Printf("Got beacon: " + e.Beacon.Hwid)
		}
	}
}

// GetImageBinary: Get image binary from LINE server based on message ID.
func GetImageBinary(blob *messaging_api.MessagingApiBlobAPI, messageID string) ([]byte, error) {
	// Get image binary from LINE server based on message ID.
	content, err := blob.GetMessageContent(messageID)
	if err != nil {
		log.Println("Got GetMessageContent err:", err)
	}
	defer content.Body.Close()
	data, err := io.ReadAll(content.Body)
	if err != nil {
		log.Fatal(err)
	}

	return data, nil
}
