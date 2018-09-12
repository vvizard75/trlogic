IMAGE := vvizard/trlogic

test:
	go test -v ./...

image:
	docker build -t $(IMAGE) .

push-image:
	docker login -u $(DOCKER_USER) -p $(DOCKER_PASS)
	export TAG=`if [ "$(TRAVIS_BRANCH)" == "master" ]; then echo "latest"; else echo "$(TRAVIS_BRANCH)" ; fi`
	docker build -f Dockerfile -t $(REPO):$(COMMIT) .
	docker tag $(IMAGE):$(COMMIT) $(IMAGE):$(TAG)
	docker tag $(IMAGE):$(COMMIT) $(IMAGE):travis-$(TRAVIS_BUILD_NUMBER)
	docker push $(IMAGE)