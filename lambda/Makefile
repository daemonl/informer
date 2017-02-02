FUNC_NAME=IntactUpload

runtime:
	go build -o runtime ../

lambda.zip: runtime
	rm -f lambda.zip
	zip -r lambda.zip index.js runtime

aws/upload: lambda.zip
	aws s3 cp lambda.zip s3://lambda.tpt/$(FUNC_NAME).zip

aws/update: aws/upload
	aws lambda update-function-code \
		--function-name $(FUNC_NAME) \
		--s3-bucket lambda.tpt \
		--s3-key $(FUNC_NAME).zip

aws/create: aws/upload
	aws lambda create-function \
		--function-name $(FUNC_NAME) \
		--role arn:aws:iam::543266951854:role/ApexLambda \
		--runtime nodejs4.3 \
		--handler index.handler \
		--code S3Bucket=lambda.tpt,S3Key=$(FUNC_NAME).zip

.PHONY: runtime
