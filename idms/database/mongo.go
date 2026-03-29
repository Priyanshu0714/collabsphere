package database

import (
	"backend/idms/services"
	"context"
	"os"
	"time"
	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FileInformation struct {
	Etag string `json:"etag" bson:"etag"`
	Filename string `json:"name" bson:"name"`
	LastModified time.Time `json:"lastModified" bson:"lastModified"`
	Size int64 `json:"size" bson:"size"`
	ContentType string `json:"contentType" bson:"contentType"`
	AIData services.Data `json:"aiData" bson:"aiData"`
}

func MongoDB (result services.Data,fileData minio.ObjectInfo) error{

	uri:=os.Getenv("MONGO_URL")
	client,err:=mongo.Connect(options.Client().ApplyURI(uri))

	if err!=nil{
		return err
	}

	// for closing the connection
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll:=client.Database("CollabSphere").Collection("FilesData")
	doc:=FileInformation{
		Etag: fileData.ETag,
		Filename: fileData.Key,
		LastModified: fileData.LastModified,
		Size: fileData.Size,
		ContentType: fileData.ContentType,
		AIData: result,
	}
	_,err=coll.InsertOne(context.TODO(),doc)
	return nil
}