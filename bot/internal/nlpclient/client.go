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
type chatPlusResp struct {
	IntentSlug       string  `json:"intent_slug"`
	IntentConfidence float64 `json:"intent_confidence"`
	MiniAnswer       *string `json:"mini_answer"`
	LLMAnswer        string  `json:"llm_answer"`
}

func (c *Client) ChatPlus(text, lang string, history []map[string]string) (mini *string, llm string, err error) {
	sys := map[string]string{"role": "system", "content": systemForLang(lang)}
	msgs := append([]map[string]string{sys}, history...)
	body, _ := json.Marshal(struct {
		Text    string              `json:"text"`
		History []map[string]string `json:"history,omitempty"`
	}{Text: text, History: msgs})

	req, err := http.NewRequest("POST", c.Base+"/chat_plus", bytes.NewBuffer(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HC.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var out chatPlusResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, "", err
	}
	return out.MiniAnswer, out.LLMAnswer, nil
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
		return "Сен TalapkerBot WKATU көмекшісісің. Пайдаланушы қай тілде жазса, сол тілде қысқа да нақты жауап бер."
	}
	return "Ты помощник TalapkerBot WKATU. Отвечай кратко и на том же языке, что и пользователь."
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
