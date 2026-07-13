# glazius

### Часть 1. Техническое Задание (ТЗ) на MVP.

**1. Общие сведения**
*   **Имя проекта:** `glazius` (CLI-утилита)
*   **Платформа и стек:** macOS (darwin/arm64), Go 1.26.4+. Без использования легаси-решений.
*   **Архитектура:** Монолит на базе принципов DDD и чистой архитектуры (Ports and Adapters). В будущем планируется вынесение ядра в демона (Linux) и добавление Telegram/HTMX интерфейсов.
*   **Методология:** Strict TDD (Сначала пишем тесты на домен и юзкейсы, затем реализацию).
*   **Хранение данных:**
    *   Локальный JSON-файл (база данных для отслеживаемых сериалов, кредов и кук).
    *   Локальная скрытая папка-кеш для скачанных `.torrent` файлов.
*   **Сторонние зависимости (строго в слое Infrastructure):**
    *   `github.com/PuerkitoBio/goquery` (для парсинга HTML).
    *   `github.com/zeebo/bencode` (или аналог для парсинга Bencode торрент-файлов).

**2. Основные бизнес-сценарии (CLI-команды)**
1.  **`glazius login <user> <pass>`**
    *   Сохраняет учетные данные в JSON (в MVP — plain-text).
    *   Выполняет авторизацию на rutracker.org, получает сессионную куку и сохраняет её. В дальнейшем при протухании куки приложение должно автоматически и незаметно перелогиниваться.
2.  **`glazius add <url>`**
    *   Получает страницу по ссылке, извлекает название раздачи и текущий `info_hash`.
    *   Скачивает `.torrent` файл в локальный кеш.
    *   Добавляет сериал в JSON базу (статус: отслеживается, изменений нет).
3.  **`glazius list`**
    *   Выводит список отслеживаемых сериалов: ID, Название, Наличие обновлений.
4.  **`glazius remove <id>`**
    *   Удаляет запись из JSON базы и удаляет связанные `.torrent` файлы из кеша.
5.  **`glazius check`**
    *   Обходит все сериалы из базы. Парсит страницу, сравнивает `info_hash` с последним подтвержденным хешем из базы.
    *   Если хеш изменился: скачивает новый `.torrent`, парсит его и старый `.torrent`, вычисляет разницу (новые файлы).
    *   Помечает раздачу в базе флагом `PendingAck` (ожидает подтверждения) и сохраняет новый хеш как "кандидата".
    *   Выводит на экран сообщение: *"Обновился сериал [Название]. Добавлены файлы: ..."*. Если вызвать `check` повторно до подтверждения, снова выводит это уведомление без повторного скачивания.
6.  **`glazius ack <id>`**
    *   Снимает флаг `PendingAck` для указанного сериала, делает новый хеш "базовым".
    *   Копирует актуальный `.torrent` файл из локального кеша в **текущую директорию** (откуда запущена команда) и выводит путь к сохраненному файлу.

### Часть 2. Архитектурный план (DDD / Clean Architecture)

Чтобы проект можно было легко масштабировать в микросервисы и прикручивать к нему Telegram-бота, мы жестко разделим код на слои. Доменный слой не будет знать ничего о JSON, CLI, HTTP или HTML.

#### Структура директорий проекта:
```text
glazius/
├── cmd/
│   └── glazius/
│       └── main.go                 # Точка входа, сборка зависимостей (Dependency Injection), парсинг CLI
├── internal/
│   ├── domain/                     # 1. ДОМЕННЫЙ СЛОЙ (Бизнес-логика и интерфейсы)
│   │   ├── entity/                 # Сущности (Series, TorrentInfo, User)
│   │   ├── vo/                     # Value Objects (URL, InfoHash, FileDiff)
│   │   ├── service/                # Доменные сервисы (например, логика сравнения файлов)
│   │   └── port/                   # Интерфейсы для инфраструктуры (Repositories, Clients, Parsers)
│   ├── application/                # 2. СЛОЙ ПРИЛОЖЕНИЯ (Use Cases / Сценарии)
│   │   ├── usecase/                # Координация слоев (TrackUseCase, CheckUseCase, AuthUseCase)
│   ├── infrastructure/             # 3. ИНФРАСТРУКТУРНЫЙ СЛОЙ (Реализация портов)
│   │   ├── trackerclient/          # Реализация парсинга rutracker через net/http + goquery
│   │   ├── torrentparser/          # Реализация парсинга .torrent через bencode
│   │   └── storage/                # Реализация работы с JSON-БД и файловым кешем
│   └── presentation/               # 4. СЛОЙ ПРЕДСТАВЛЕНИЯ (CLI)
│       └── cli/                    # Обработчики команд терминала
├── go.mod
└── go.sum
```

#### Детализация слоев (Проектирование):

**1. Domain (Домен)**
Самый независимый слой. Только чистый Go.
*   **Сущности (Entities):**
    *   `User`: Содержит логин, пароль, текущую куку (строку).
    *   `Series`: Отслеживаемая раздача. Поля: `ID`, `URL`, `Title`, `BaseInfoHash` (последний подтвержденный), `LatestInfoHash` (актуальный на трекере), `PendingAck` (bool).
    *   `TorrentData`: Внутренняя структура торрента (список файлов `TorrentFileItem` с именами и размерами).
*   **Порты (Ports) — интерфейсы, которые домен требует от внешнего мира:**
    *   `DataRepository`: `SaveUser()`, `GetUser()`, `SaveSeries()`, `ListSeries()`, `GetSeries(id)`, `DeleteSeries(id)`.
    *   `TorrentStorage`: `SaveToCache(hash, bytes)`, `GetFromCache(hash)`, `CopyTo(hash, destPath)`, `RemoveFromCache(hash)`.
    *   `TrackerClient`: `Login(user, pass) -> cookie`, `FetchSeriesInfo(url, cookie) -> (Title, InfoHash)`, `DownloadTorrent(url, cookie) -> bytes`.
    *   `TorrentParser`: `Parse(bytes) -> TorrentData`.
*   **Доменные сервисы (Domain Services):**
    *   `TorrentDiffService`: Принимает два `TorrentData` (старый и новый) и возвращает список добавленных файлов.

**2. Application (Слой приложения / Use Cases)**
Здесь мы реализуем наши команды. Они принимают интерфейсы (через DI).
*   `AuthService`: Получает креды, вызывает `TrackerClient.Login`, сохраняет в `DataRepository`.
*   `TrackService`: Метод `Add(url)` — использует Client для парсинга инфы, качает торрент, сохраняет в Storage и Repo. Метод `Remove(id)`, Метод `List()`.
*   `CheckerService`: Метод `CheckUpdates()` — реализует бизнес-логику (п. 5 из ТЗ). Оркестрирует вызовы Client, Repo, TorrentStorage, TorrentParser и TorrentDiffService.
*   `AckService`: Метод `Acknowledge(id)` — меняет статус в Repo, дергает `TorrentStorage.CopyTo`.

**3. Infrastructure (Адаптеры)**
Здесь живет "грязный" код.
*   `storage.JSONRepository` — читает/пишет `data.json`.
*   `storage.FileCache` — пишет бинарники в `~/.glazius/cache/`.
*   `trackerclient.Rutracker` — делает HTTP-запросы. Здесь живет `goquery`. Если кука протухла (HTTP 302 на страницу логина или нет нужных селекторов), он делает повторный логин.
*   `torrentparser.Bencode` — использует библиотеку Bencode для разбора бинарников.

**4. Presentation (CLI)**
Парсинг аргументов командной строки (`os.Args` или стандартный пакет `flag`). Вызовы соответствующих методов из слоя Application и красивый вывод результатов (Print) в консоль.
