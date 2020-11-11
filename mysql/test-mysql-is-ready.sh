#!/bin/bash
set -e

ATTEMPTS=0

until docker exec mysql-volumeless mysql -h localhost -u moov -psecret --protocol=TCP -e "SELECT VERSION();SELECT NOW()" test || [ $ATTEMPTS -ge 10 ]
do
	((ATTEMPTS+=1))
	echo "Waiting for database connection... ($ATTEMPTS)"
	# wait for 5 seconds before check again
	sleep 3
done
