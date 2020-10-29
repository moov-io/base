build:
	docker build -t mysql-volumeless:8.0 .

run:
	docker run --name mysql-volumeless \
		-e MYSQL_USER=moov \
		-e MYSQL_PASSWORD=secret \
		-e MYSQL_ROOT_PASSWORD=secret \
	       	mysql-volumeless:8.0

stop:
	docker stop mysql-volumeless

image:
	docker commit mysql-volumeless mysql-volumeless:8.0
	docker tag mysql-volumeless:8.0 vaulty/mysql-volumeless:8.0

push:
	docker push vaulty/mysql-volumeless:8.0
