package handler

import (
	"backend/idms/services"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type QueryHandler struct{
	Minio *services.MinIOService
	Mongo *mongo.Client
}

func(h *QueryHandler) QueryProcessing(c echo.Context)  error{
	query:=c.FormValue("query")
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
	Text:%s`, query)

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
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var groqResp services.GroqResponse
	err = json.Unmarshal(respBody, &groqResp)
	if err != nil {
		return err
	}
	content:=groqResp.Choices[0].Message.Content

	start:=strings.Index(content,"{")
	end:=strings.LastIndex(content,"}")+1

	CleanData:=content[start:end]
	var result services.Data
	err =json.Unmarshal([]byte(CleanData),&result)
	if err!=nil{
		return err
	}
	coll:=h.Mongo.Database("CollabSphere").Collection("FilesData")
	// now finding the data in mongodb 
	// coll:=client.Database.
	orfilter:=[]bson.M{}
	if result.Title!="" && result.Title!="null"{
		orfilter = append(orfilter, bson.M{
			"aiData.title":result.Title,
		})
	}
	if result.Company!="" && result.Company!="null"{
		orfilter=append(orfilter, bson.M{
			"aiData.company":result.Company,
		})
	}
	if result.Date!="" && result.Date!="null"{
		orfilter = append(orfilter, bson.M{
			"aiData.date":result.Date,
		})
	}
	if result.People!=nil{
		orfilter = append(orfilter, bson.M{
			"aiData.people":result.People,
		})
	}
	if result.Location!=nil{
		orfilter = append(orfilter, bson.M{
			"aiData.location":result.Location,
		})
	}
	if result.Financials.Profit!="null"{
		orfilter = append(orfilter, bson.M{
			"aiData.financials.profit":result.Financials.Profit,
		})
	}
	if result.Financials.Revenue!="null"{
		orfilter = append(orfilter, bson.M{
			"aiData.financials.revenue":result.Financials.Revenue,
		})
	}
	filter:=bson.M{}
	if len(orfilter)>0{
		filter["$or"]=orfilter
	}
	// final query
	cursor,err:=coll.Find(context.TODO(),filter)
	if err!=nil{
		return err
	}
	return c.JSON(http.StatusOK,cursor)
}