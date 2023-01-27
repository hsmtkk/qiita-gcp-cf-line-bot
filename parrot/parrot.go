package parrot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("parrot", parrot)
}

type MessageEvents struct {
	Events []MessageEvent `json:"events"`
}

type MessageEvent struct {
	ReplyToken string  `json:"replyToken"`
	Message    Message `json:"message"`
}

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ReplyMessage struct {
	ReplyToken string    `json:"replyToken"`
	Messages   []Message `json:"messages"`
}

func parrot(w http.ResponseWriter, r *http.Request) {
	dumpReq, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatalf("failed to dump HTTP request; %v", err.Error())
	}
	log.Print("dump HTTP request")
	log.Print(string(dumpReq))

	events, err := receive(r)
	if err != nil {
		log.Fatal(err)
	}

	if len(events.Events) == 0 {
		log.Print("no event")
	} else {
		if err := reply(events.Events[0].ReplyToken, events.Events[0].Message.Text); err != nil {
			log.Fatal(err)
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Fatalf("failed to write response; %v", err.Error())
	}
}

func receive(r *http.Request) (MessageEvents, error) {
	var events MessageEvents
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return events, fmt.Errorf("failed to read request body; %w", err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(reqBytes, &events); err != nil {
		return events, fmt.Errorf("failed to decode JSON; %w", err)
	}

	return events, nil
}

func reply(replyToken, text string) error {
	url := "https://api.line.me/v2/bot/message/reply"
	body, err := json.Marshal(ReplyMessage{
		ReplyToken: replyToken,
		Messages:   []Message{{Type: "text", Text: text}},
	})
	if err != nil {
		return fmt.Errorf("failed to encode JSON; %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to make new request; %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("CHANNEL_ACCESS_TOKEN"))

	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return fmt.Errorf("failed to dump HTTP request; %w", err)
	}
	log.Print("dump HTTP request")
	log.Print(string(reqDump))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request; %w", err)
	}
	defer resp.Body.Close()

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Errorf("failed to dump HTTP response; %w", err)
	}
	log.Print("dump HTTP response")
	log.Print(string(respDump))

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("got error HTTP response; %d; %s", resp.StatusCode, resp.Status)
	}
	return nil
}
