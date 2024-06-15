STACK_NAME ?= itsa-api
GO := go
USER_FUNCTIONS := get-users get-user create-user update-user delete-user
POINT_FUNCTIONS := get-points create-points update-points delete-points
MAKER_FUNCTIONS := get-transactions create-transaction update-transaction
ROLE_FUNCTIONS := get-role get-roles create-role update-role delete-role
LOG_FUNCTIONS := get-logs
ADMIN_FUNCTIONS := authorizer
REGION := ap-southeast-1

build-user:
	${MAKE} ${MAKEOPTS} $(foreach userFunction,${USER_FUNCTIONS}, build-user-${userFunction})

build-user-%:
	cd functions/users/$* && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build -o bootstrap

build-point:
	${MAKE} ${MAKEOPTS} $(foreach pointFunction,${POINT_FUNCTIONS}, build-point-${pointFunction})

build-point-%:
	cd functions/points/$* && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build -o bootstrap

build-maker:
	${MAKE} ${MAKEOPTS} $(foreach makerFunction,${MAKER_FUNCTIONS}, build-maker-${makerFunction})

build-maker-%:
	cd functions/maker-checker/$* && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build -o bootstrap

build-role:
	${MAKE} ${MAKEOPTS} $(foreach roleFunction,${ROLE_FUNCTIONS}, build-role-${roleFunction})

build-role-%:
	cd functions/roles/$* && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build -o bootstrap

build-logs:
	${MAKE} ${MAKEOPTS} $(foreach logFunction,${LOG_FUNCTIONS}, build-logs-${logFunction})

build-logs-%:
	cd functions/logs/$* && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build -o bootstrap

build-administrative:
	${MAKE} ${MAKEOPTS} $(foreach adminFunction,${ADMIN_FUNCTIONS}, build-administrative-${adminFunction})

build-administrative-%:
	cd functions/admin/$* && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 ${GO} build -o bootstrap

build: build-user build-point build-maker build-role build-logs build-administrative

clean:
	@rm $(foreach function,${USER_FUNCTIONS}, functions/users/${function}/bootstrap)
	@rm $(foreach function,${POINT_FUNCTIONS}, functions/points/${function}/bootstrap)
	@rm $(foreach function,${MAKER_FUNCTIONS}, functions/maker-checker/${function}/bootstrap)
	@rm $(foreach function,${ROLE_FUNCTIONS}, functions/roles/${function}/bootstrap)
	@rm $(foreach function,${ADMIN_FUNCTIONS}, functions/admin/${function}/bootstrap)

deploy:
	sam build && sam deploy --stack-name ${STACK_NAME};

deploy-auto: 
	@sam deploy --stack-name ${STACK_NAME} --no-confirm-changeset --no-fail-on-empty-changeset;

deploy-full-auto: build-user build-point build-maker build-role build-logs build-administrative
	@sam deploy --stack-name ${STACK_NAME} --no-confirm-changeset --no-fail-on-empty-changeset;

delete:
	@sam delete --stack-name ${STACK_NAME}

build-run:
	sam build && DOCKER_HOST=unix://${HOME}/.docker/run/docker.sock sam local start-api --env-vars env.dev.json

teardown:
	sam delete --stack-name ${STACK_NAME}

deploy-jj:
	sam deploy --stack-name ${STACK_NAME} --profile itsa

db-reset:
	go run seed/main.go
