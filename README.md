# repeater-ait

Проект представляет собой UDP MPEG-TS репитер, который добавляет в поток PID, содержащий AIT (Application Information Table) с HbbTV-ссылкой. Это полезно для сценариев, связанных с гибридным вещанием (телевидение + интернет), особенно для реализации функций HbbTV.

### Основное назначение

- **Повторение потока:** Устройство или программа принимает UDP MPEG-TS поток и ретранслирует его на другой адрес или порт.
- **Добавление AIT:** Программа добавляет в поток новый PID (обычно 710), содержащий Application Information Table.
- **HbbTV-приложение** (например, интерактивный интерфейс или онлайн-сервис).

### Для чего это нужно?

HbbTV позволяет добавить к традиционному телевидению интернет-контент, такой как приложения, телегиды, рекламные или информационные панели.
AIT является ключевой частью HbbTV-архитектуры, так как она указывает устройствам (например, смарт-ТВ), как и куда обращаться за интерактивным контентом.

Тестирование и разработка:

- Удобный инструмент для тестирования или разработки HbbTV-приложений.
- Позволяет легко модифицировать поток без необходимости изменения конфигурации вещательной аппаратуры.

### Обеспечение совместимости:

Для разработчиков и интеграторов, работающих с HbbTV, репитер упрощает добавление AIT в существующие потоки.

### Применение

- **Операторы вещания:** Для добавления интерактивного контента в эфир.
- **Тестовые стенды:** Для проверки работы смарт-ТВ или других устройств с поддержкой HbbTV.
- **Разработчики:** Для генерации потоков с AIT при создании приложений.

### Возможности проекта

- Программная реализация позволяет быстро разворачивать и модифицировать.
- Поддержка различных UDP-потоков, включая multicast.
- Возможность настройки PID и содержимого AIT в исходном коде.

## compile
`go build -ldflags "-linkmode external -extldflags '-static'" -o repeater-ait`

## usage
`./repeater-ait <udp:port> <hbblink>`

## example
`./repeater-ait udp://eth1@239.0.100.1:1234 http://hbbtv.com/app`


![Screenshot cascap](https://github.com/unidiag/repeater-ait/blob/main/screenshot.jpg)