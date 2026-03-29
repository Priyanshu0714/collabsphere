package services

import (
	"io"
	"net/http"
	"os"
)

func TextExtractor(location string) (string, error){

	final_location:="/tmp/"+location
	file,err:=os.Open(final_location)
	if err!=nil {
		return " ",err
	}
	// creating a new http client request
	req, err:=http.NewRequest("PUT","http://localhost:9998/tika",file)

	if err != nil {
		return "",err
	}
	// now setting the headers as we want plain text
	req.Header.Add("Accept", "text/plain")

	// now executing request
	response,err:=http.DefaultClient.Do(req)
	if err !=nil{
		return "Some Error Occured during text extraction", err
	}
	defer response.Body.Close()

	body,_:=io.ReadAll(response.Body)

	return string(body),nil

}