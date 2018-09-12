IMAGE := vvizard/trlogic

test:
	go test -v ./...

image:
	docker build -t $(IMAGE) .

push-image:
	docker login -u $(DOCKER_USER) -p $(DOCKER_PASS)
	echo $(TAG)
	docker build -f Dockerfile -t $(IMAGE):$(COMMIT) .
	docker tag $(IMAGE):$(COMMIT) $(IMAGE):$(TAG)
	docker tag $(IMAGE):$(COMMIT) $(IMAGE):travis-$(TRAVIS_BUILD_NUMBER)
	docker push $(IMAGE)