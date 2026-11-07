package handler

import (
	"backend/idms/database"
	"backend/idms/services"
	"fmt"

	"github.com/dutchcoders/go-clamd"
	"github.com/labstack/echo/v4"

	// "golang.org/x/tools/go/types/objectpath"

	// "net/http"p
	"strings"
	// "time"
	// "errors"
	"os"
)

type ProcessHandler struct {
	Minio *services.MinIOService
}

func (h *ProcessHandler) PostProcessFile(c echo.Context) error {
	username := getUsername(c)
	temppath:=fmt.Sprintf("%s/%s",username,strings.TrimSpace(c.QueryParam("filename")))
	fmt.Print(temppath)
	fmt.Print("\n\n")
	l_path,err:=h.Minio.DownloadFile(temppath)
	// fmt.Print("This is the context request - ",c.Request().Context())
	// fmt.Print(l_path,err)
	// l_path := "/tmp/priyanshu/1774421013_testing2"
	
	file, err := os.Open(l_path)
	if err != nil {
		return c.JSON(500, err.Error())
	}
	defer file.Close()
	clamav := clamd.NewClamd("tcp://127.0.0.1:3310")
	response, err := clamav.ScanStream(file, make(chan bool))
	if err != nil {
		return c.JSON(500, err.Error())
	}
	for r := range response {
		fmt.Println("path: ",r.Path)
		fmt.Println("description: ",r.Description)
		fmt.Println("Scan result:", r.Status)
		fmt.Print(clamd.RES_FOUND)
		fmt.Print(clamd.RES_ERROR)
		fmt.Print(clamd.RES_OK)

		if r.Status == clamd.RES_OK {
			fmt.Print("\n\n......................FILE OKKKK......................\n\n")
			// moving file from the temp bucket to main bucket
			// first we get the file information
			fileData,err:=h.Minio.GetFileInfo("tempbucket",temppath)
			if err != nil {
				return c.JSON(500, map[string]string{"error": "Failed to get file info: " + err.Error()})
			}
			// now first we need to extract the text if type == pdf else we dont need to extract.
			if fileData.ContentType == "application/pdf"{
				extractedText, err := services.TextExtractor(temppath)
				if err != nil {
					return c.JSON(500, map[string]string{"error": "Text extraction failed: " + err.Error()})
				}
				Meta_Header, err := services.AIHeader(extractedText)
				if err != nil {
					return c.JSON(500, map[string]string{"error": "AI header extraction failed: " + err.Error()})
				}
				fmt.Print(Meta_Header)

				// now before moving the file to the main bucket, we will add the details in our mongodb database
				if err := database.MongoDB(Meta_Header, fileData); err != nil {
					return c.JSON(500, map[string]string{"error": "MongoDB insert failed: " + err.Error()})
				}

			}
			// now we need to store all the data in the mongodb database
			h.Minio.MoveFile(c.Request().Context(),temppath)
			return c.JSON(200, fileData)
		}
		if r.Status == clamd.RES_FOUND {
			h.Minio.DeleteFile(c.Request().Context(),"tempbucket",temppath)
			return c.String(200, "Virus detected")
		}
	}
	return c.String(400,"Client Side Error Occured")
}
