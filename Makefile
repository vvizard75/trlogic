IMAGE := vvizard/trlogic

test:
	go test -v ./...

image:
	docker build -t $(IMAGE) .

push-image:
	docker push $(IMAGE)
	docker login -e $DOCKER_EMAIL -u $DOCKER_USER -p $DOCKER_PASS
	docker push vvizard/trlogic