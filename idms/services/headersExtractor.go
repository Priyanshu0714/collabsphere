package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Financial struct{
	Revenue string `json:"revenue"`
	Profit string	`json:"profit"`
}

type Data struct{
	Title string `json:"title"`
	Company string `json:"company"`
	Date string `json:"date"`
	People []string `json:"people"`
	Location []string `json:"location"`
	Financials Financial`json:"financials"`
}
type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func AIHeader(txt string) (Data, error) {

	url := "https://api.groq.com/openai/v1/chat/completions"
	prompt := fmt.Sprintf(`Return ONLY valid JSON in this exact format:
	{
		"title": "",
		"company": "",
		"date": "",
		"people": [],
		"location": [],
		"financials": {
			"revenue": "",
			"profit": ""
		}
	}
	Rules:
	- If a field is not present, return null
	- Do not add extra text
	- Do not explain anything
	- Only JSON
	Text:%s`, txt)

	body := map[string]interface{}{
		"model": "llama-3.1-8b-instant",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.2,
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))

	req.Header.Set("Authorization", "Bearer "+os.Getenv("API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Data{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var groqResp GroqResponse
	err = json.Unmarshal(respBody, &groqResp)
	if err != nil {
		return Data{}, err
	}
	content:=groqResp.Choices[0].Message.Content

	start:=strings.Index(content,"{")
	end:=strings.Index(content,"}")

	CleanData:=content[start:end]
	var result Data
	err =json.Unmarshal([]byte(CleanData),&result)
	if err!=nil{
		return Data{},err
	}
	return result, nil
}
