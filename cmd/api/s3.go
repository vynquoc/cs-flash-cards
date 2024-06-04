package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (app *application) uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("image")
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	defer file.Close()
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	fileName := handler.Filename
	_, err = app.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(app.config.s3.bucketName),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(buf.Bytes()),
		ACL:    aws.String("public-read"),
	})

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	imageURL := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", app.config.s3.bucketName, app.config.s3.region, fileName)
	err = app.writeJSON(w, http.StatusCreated, envelope{"image_url": imageURL}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
