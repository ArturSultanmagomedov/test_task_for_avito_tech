install_postgres:
	docker pull postgres

create_db:
	docker run --name=avito-tech-db -e POSTGRES_PASSWORD='1234' -p 5433:5432 -d --rm postgres


