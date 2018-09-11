IMAGE := vvizard/trlogic

test:
    go test -v ./...

image:
	docker build -t $(IMAGE) .

push-image:
	docker push $(IMAGE)