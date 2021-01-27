start:
	docker-compose down && docker-compose up --build

restart:
	docker-compose down && docker-compose up

stop:
	docker-compose down