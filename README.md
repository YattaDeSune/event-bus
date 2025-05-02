<a id="about"></a>
## About 👀
Сервис подписок, работающий по **gRPC**, основанный на шине событий по принципу **Publisher-Subscriber**.

## Содержание 📜
- [About](#about)
- [Структура проекта](#структура-проекта)
- [How it works](#how-it-works)
- [Переменные окружения](#переменные-окружения)
- [Quick start](#quick-start)
- [Example](#example)
- [Другие особенности проекта](#другие-особенности-проекта)
- [Контакты](#contacts)

<a id="структура-проекта"></a>
## Структура проекта 🚧
Структура составлена в соответствии с `project layout`:
```
сmd
  - server
    -- main.go         // точка входа сервиса
internal
  - config
    -- config.go       // получение конфига из .env
  - logger
    -- zap_logger.go   // сборка логгера
  - proto
    -- subpub.proto    // gRPC методы
    -- сгенерированные файлы
  - server
    -- server.go       // сборка сервера
  - service
    -- service.go      // сборка gRPC-сервиса
pkg
  - subpub
    -- subpub.go       // логика шины
    -- subpub_test.go  // тесты для пакета subpub
.env       // переменные окружения
Makefile   // команды для запуска и тестирования
```

<a id="how-it-works"></a>
## How it works 🎯
Сервис предоставляет два метода:
- `Subscribe` - подписка на события по ключу
- `Publish` - публикация события по ключу
У одного события может быть несколько слушателей

По умолчанию сервер запускается на порту `:50051`.

Создание подписки:
```
message SubscribeRequest {
    string key = 1;
}
```

Публикация события:
```
message PublishRequest {
    string key = 1;
    string data = 2;
}
```

Ответ слушателя:
```
message Event {
    string data = 1;
}
```

<a id="переменные-окружения"></a>
## Переменные окружения 📩
Конфиг реализован с помощью переменных окружения из файла `.env`

```env
HOST="0.0.0.0"          // хост сервера
PORT="50051"            // порт сервера
SHUTDOWN_TIMEOUT_S=5s   // таймаут для gracefu; shutdown
MAX_CONN_IDLE_S=300s    // время бездействия слушателя
```

<a id="quick-start"></a>
## Quick start ⚡
В проекте присутствует `Makefile`, с его помощью можно запустить проект и протестировать пакет `subpub`.

- Запуск сервера:
```
make run
```
Под капотом - go run cmd/server/main.go

- Запуск тестов: (из корня проекта)
```
make test
```
// под капотом - go test -v ./pkg/subpub/...

Далее рассмотрим пример работы сервиса.

<a id="example"></a>
## Example 🔴
Запуск будем производить из корня проекта.

1. В первом терминале запустим сервер:
```
make run
```

2. Во втором терминале создадим подписку на событие `test`:
`grpcurl -proto internal/proto/subpub.proto -plaintext -d '{"key": "test"}' localhost:50051 pubsub.PubSub/Subscribe`

3. В третьем терминале создадим событие `test`:
`grpcurl -proto internal/proto/subpub.proto -plaintext -d '{"key": "test", "data": "test message"}' localhost:50051 pubsub.PubSub/Publish`

В терминале слушателя увидим:
```
{
  "data": "test message"
}
```

А в логах первого терминала увидим все события - от создания сервера до публикации событий и подписок.

<a id="другие-особенности-проекта"></a>
## Другие особенности проекта 🐯
- Подписки в пакете `subpub` обрабатываются асинхронно: медленные подписчики не тормозят других
- С помощью каналов сохраняется порядок сообщений (FIFO)
- Логгирование в проекте реализованно с помощью логгера zap
- для сервера реализован `graceful shutdown`
- dependency injection и конструкторы: в main создаются экземпляры сервера, логгера и конфига и передаются вниз по цепочке

<a id="contacts"></a>
## Contacts 💬
<div id="contacts">
  <a href="https://t.me/YattaDesuNe">
    <img src="https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white" alt="Telegram Badge"/>
  </a>
  <a href="mailto:belyaevlv742@gmail.com">
    <img src="https://img.shields.io/badge/Gmail-D14836?style=for-the-badge&logo=gmail&logoColor=white" alt="Email Badge"/>
  </a>
</div>
<img src="https://komarev.com/ghpvc/?username=YattaDeSune&style=flat-square&color=blue" alt=""/>
