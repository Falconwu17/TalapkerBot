package nlpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	Base string
	HC   *http.Client
}

type askReq struct {
	Text string `json:"text"`
}
type askResp struct {
	Slug       string  `json:"slug"`
	Confidence float64 `json:"confidence"`
	BestPhrase string  `json:"best_phrase"`
}

func New(base string) *Client {
	return &Client{
		Base: base,
		HC:   &http.Client{Timeout: 20 * time.Second},
	}
}

type chatReq struct {
	Text    string              `json:"text"`
	History []map[string]string `json:"history,omitempty"`
}
type chatResp struct {
	Answer string `json:"answer"`
}

func (c *Client) Classify(text string) (slug string, conf float64, err error) {
	body, _ := json.Marshal(askReq{Text: text})
	req, err := http.NewRequest("POST", c.Base+"/ask", bytes.NewBuffer(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("nlp status %d", resp.StatusCode)
	}
	var out askResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", 0, err
	}
	return out.Slug, out.Confidence, nil
}

func systemForLang(lang string) string {
	if lang == "kz" {
		return "Сен TalapkerBot WKATU көмекшісісің. ҚОЛДАНУШЫ ҚАЗАҚША ЖАЗСА — ҚАЗАҚША ЖАУАП БЕР. Қысқа және нақты жауап бер. Университет жайлы факті ойдан қоспа."
	}
	return "Ты помощник TalapkerBot WKATU. ОТВЕЧАЙ НА ТОМ ЖЕ ЯЗЫКЕ, ЧТО И ПОЛЬЗОВАТЕЛЬ (рус/каз). Отвечай кратко и по делу. Не выдумывай факты об университете."
}

func (c *Client) Chat(text, lang string, history []map[string]string) (string, error) {
	sys := map[string]string{"role": "system", "content": systemForLang(lang)}
	msgs := append([]map[string]string{sys}, history...)

	payload := chatReq{
		Text:    text,
		History: msgs,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", c.Base+"/chat", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nlp chat status %d", resp.StatusCode)
	}
	var out chatResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Answer, nil
}
