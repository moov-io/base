DOCKER_CONTAINER_LIST := $(shell docker ps -aq -f status=exited -f name=mysql-volumeless)

build: clean
	docker build -t mysql-volumeless:8.0 .
	docker run -d \
	       	--name mysql-volumeless \
		-e MYSQL_USER=moov \
		-e MYSQL_PASSWORD=secret \
		-e MYSQL_ROOT_PASSWORD=secret \
		-e MYSQL_DATABASE=test \
		-p 3306:3306 \
	       	mysql-volumeless:8.0
	./test-mysql-is-ready.sh
	docker stop mysql-volumeless
	docker commit mysql-volumeless mysql-volumeless:8.0
	docker tag mysql-volumeless:8.0 moov/mysql-volumeless:8.0
push:
	docker push moov/mysql-volumeless:8.0
clean:
	@if [ -n "$(DOCKER_CONTAINER_LIST)" ]; then docker rm "$(DOCKER_CONTAINER_LIST)"; fi;
