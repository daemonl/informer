VERSION=$(shell git describe --tags --candidates=1)
FLAGS=-X main.Version=$(VERSION)
PROJECT_NAME=informer

AWS_ACCOUNT=$(shell aws sts get-caller-identity | jq -r '.Account')
AWS_REGION=$(shell aws configure get region)
DOCKER_ROOT=$(AWS_ACCOUNT).dkr.ecr.$(AWS_REGION).amazonaws.com/$(PROJECT_NAME)

test:

docker/build: test
	docker build -t $(DOCKER_ROOT):$(VERSION) .
	docker tag $(DOCKER_ROOT):$(VERSION) $(DOCKER_ROOT):latest


ecr/exists:
	aws ecr create-repository --repository-name $(PROJECT_NAME) || echo "Already exists"

ecr/push: ecr/exists docker/build
	$(shell aws ecr get-login --no-include-email)
	docker push $(DOCKER_ROOT):$(VERSION)
	docker push $(DOCKER_ROOT):latest

cf/%: ecr/push
	aws cloudformation $*-stack \
		--stack-name $(PROJECT_NAME) \
		--template-body file://./cloudformation.yml \
		--parameters ParameterKey=DockerImage,ParameterValue=$(DOCKER_ROOT):$(VERSION) \
		--capabilities CAPABILITY_IAM

