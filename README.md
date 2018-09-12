# Image gallery Web-Service

Simple web service for image gallery

## Table of Contents
1. [Installation](#installation)
2. [Docker](#docker)
3. [API](#api)

## Installation
    git clone https://github.com/vvizard75/trlogic
    cd trlogic
    go build
    ./trlogic

## Docker
    docker run -p 8081:8081 vvizard/trlogic
And get access http://localhost:8081

You can run service with volumes for store file

    git clone https://github.com/vvizard75/trlogic
    cd trlogic
    docker-compose up

## API

You must use POST request to /upload

- you can send a multipart request looks a little like this:


    POST /upload HTTP/1.1
    Host: localhost:8081
    Content-Type: multipart/form-data; boundary=MultipartBoundry
    Accept-Encoding: gzip, deflate

    --MultipartBoundry
    Content-Disposition: form-data; name="files"; filename="bestimagen.jpg"
    Content-Type: image/jpeg

    rawimagecontentwhichlooksfunnyandgoesonforever.d.sf.d.f.sd.fsdkfjkslhfdkshfkjsdfdkfh
    --MultipartBoundry--

- you can send a request with URL images from Internet looks a little like this:


    POST /upload HTTP/1.1
    Host: localhost:8081
    Content-Type: application/x-www-form-urlencoded;
    path=https://cdn.pixabay.com/photo/2015/10/06/19/28/landscape-975091_960_720.jpg

- you can send a request with image in BASE64 looks a little like this:


	POST /upload HTTP/1.1
	Host: localhost:8080
	Content-Type: application/json
	{
  	"name": "test_image",
  	"img_data": "/9j/4RRtRXhpZgAATU0AKgAAAAgABwESAAMAAAABAAEAAAEaAAUAAAABAAAAYgEbAAUAAAABAAAAagEoAAMAAAABAAIAAAExAAIAAAAcAAAAcgEyAAIAAAAUAAAAjodpAAQAAAABAAAApAAAANAACvyAAAAnEAA"
	}