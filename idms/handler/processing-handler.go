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
	temppath:=fmt.Sprintf("%s/%s",strings.TrimSpace(c.QueryParam("username")),strings.TrimSpace(c.QueryParam("filename")))
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
		// fmt.Println("path: ",r.Path)
		// fmt.Println("description: ",r.Description)
		// fmt.Println("Scan result:", r.Status)
		// fmt.Print(clamd.RES_FOUND) result -> FOUND
		// fmt.Print(clamd.RES_ERROR)	result -> ERROR
		// fmt.Print(clamd.RES_OK)	result -> OK

		if r.Status == clamd.RES_OK {
			// moving file from the temp bucket to main bucket
			var err error
			// first we get the file information
			fileData,err:=h.Minio.GetFileInfo("tempbucket",temppath)
			// now first we need to extract the text if type == pdf else we dont need to extract.
			if fileData.ContentType == "application/pdf"{
				response,_:=services.TextExtractor(temppath)
				var Meta_Header services.Data
				Meta_Header,_=services.AIHeader(response)
				fmt.Print(Meta_Header)

				// now before moving the file to the main bucket, we will add the details in our mongodb database
				database.MongoDB(Meta_Header,fileData)

				h.Minio.MoveFile(c.Request().Context(),temppath)
			}
			// now we need to store all the data in the mongodb database
			if err != nil{
				return c.String(200,"Some error occured")
			}
			return c.JSON(200, fileData)
		}
		if r.Status == clamd.RES_FOUND {
			h.Minio.DeleteFile(c.Request().Context(),"tempbucket",temppath)
			return c.String(200, "Virus detected")
		}
	}
	return c.String(400,"Client Side Error Occured")
}
