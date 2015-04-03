EXECUTABLE = baud_server
SOURCE_FILES = $(wildcard *.go)
DEPLOY_FILES = $(EXECUTABLE) static settings.prod.json index.html
PROD = sean@worthing:deploy/baud/
PROD_ARCH = GOOS=linux GOARCH=386 CGO_ENABLED=0

.PHONY: build
build: $(EXECUTABLE)

clean:
	rm $(EXECUTABLE)
		
deploy:
	$(PROD_ARCH) go build -o $(EXECUTABLE)
	rsync -avz --delete $(DEPLOY_FILES) $(PROD)

$(EXECUTABLE): $(SOURCE_FILES)
	go build -o $(EXECUTABLE)
