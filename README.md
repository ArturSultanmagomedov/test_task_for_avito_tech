#Тестовое задание для набора на стажировку в AvitoTech (осень 2021)
Выполнил Султанмагомедов Артур

### Для начала требуется поднять базу данных (я использовал Postgres)

Вы можете поднять ее любым способом. Ниже будет представлен пример с использованием Docker.

Команды:

1. `docker pull postgres`
2. `docker run --name=<имя> -e POSTGRES_PASSWORD='<пароль>' -p <порт>:5432 -d --rm postgres`

### Теперь настроим конфиги

В папке `configs` находится файл `config.yaml` с основными настройками.

`yaml
    host: <хост сервера откуда вы будете запускать микросервис>
    port: <порт на котором сервис будет крутиться>

    db:
        username: "<юзернейм владельца базы>"
        host: "<адрес сервера где висит бд>"
        port: "<порт который используется базой>"
        dbname: "<имя базы>"
        sslmode: "<SSL мод>"`

Так же в корне проекта лежит файл `.env`. (Вообще то предполагается что его в публичном репозитории быть не должно,
но т.к мне надо обьяснить что тут вообще происходит, он здесь.)
В нем всего одно значение - пароль от бузы данных `DB_PASSWORD=<пароль>`

### Далее будут представленны описания запросов
*(Возможно с этого стоило начать, но это только возможно :) )*

---

*1. Метод начисления средств на баланс. Принимает id пользователя и сколько средств зачислить.*

формат:

    GET запрос по адресу `/add_funds`

тело запроса:

    `json { "id": <целое число>, "sum": <число дробное, строго положительное> }`

возвращает статус-код

пример запроса:
`curl --location --request GET 'localhost:8000/add_funds'
--header 'Content-Type: application/json'
--data-raw '{
"id": 1843,
"sum": 500
}'`

---

*2. Метод списания средств с баланса. Принимает id пользователя и сколько средств списать.*

формат:

    GET запрос по адресу `/write_off_funds`

тело запроса:

    `json { "id": <целое число>, "sum": <число дробное, строго положительное> }`

возвращает статус-код

пример запроса:
`curl --location --request GET 'localhost:8000/write_off_funds'
--header 'Content-Type: application/json'
--data-raw '{
"id": 177,
"sum": 3200
}'`

---

*3. Метод перевода средств от пользователя к пользователю. Принимает id пользователя с которого нужно списать средства, id пользователя которому должны зачислить средства, а также сумму.*

формат:

    GET запрос по адресу `/funds_transfer`

тело запроса:

    `json { "id1": <целое число>, "id2": <целое число>, "sum": <число дробное, строго положительное> }`

возвращает статус-код

пример запроса:
`curl --location --request GET 'localhost:8000/funds_transfer'
--header 'Content-Type: application/json'
--data-raw '{
"id1": 3,
"id2": 4,
"sum": 750
}'`

---

*4. Метод получения текущего баланса пользователя. Принимает id пользователя. Баланс всегда в рублях.*

формат:

    GET запрос по адресу `/get_balance`

тело запроса:

    `json { "id": <целое число> }`

возвращает статус-код и единственное число - значение баланса (при условии что статус-код - 200)

пример запроса:
`curl --location --request GET 'localhost:8000/get_balance'
--header 'Content-Type: application/json'
--data-raw '{
"id": 4
}'`

---

### Используемые библиотеки и фреймворки

1. Echo - фреймворк для работы с сетью
2. sqlx - библиотека для работы с реляционными базами данных
3. viper - библиотека для парсинга конфигурационных файлов
4. godotenv - библиотека для парсинга .env файлов
5. go-playground validator - библиотека для проверки валидности полей структур