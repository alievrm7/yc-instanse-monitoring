# Yandex Cloud Exporter

Экспортер для Prometheus, собирающий метрики из Yandex Cloud API.
Поддерживает несколько типов ресурсов (instances, quotas) и легко расширяется.


## Возможности

- yandex_instance_info — информация о виртуальных машинах.

- yandex_quota_usage / yandex_quota_limit — использование и лимиты квот по облаку.


## Требования

Go 1.21+

IAM-токен с правами:

- compute.viewer — для чтения виртуальных машин,

- quota-manager.viewer — для квот.


## Быстрый старт (локально)

1. Подготовьте файл с IAM-токеном, например `token.txt`:

```
<ваш_IAM_токен_одной_строкой>
```
2. Установить зависимости и собрать

```bash
# установить зависимости (один раз)
go mod tidy

# собрать в корне репозитория 
go build -o yandex-cloud-exporter ./cmd/yandex-cloud-exporter.go

```
3. Запустите экспортер:

```bash
./yandex-cloud-exporter \
  --yandex.cloud=<CLOUD_ID> \
  --yandex.token-file=./token.txt \
  --web.listen-address=":8080" \
```

4. Проверьте:

* метрики: `http://127.0.0.1:8080/metrics`
* здоровье: `http://127.0.0.1:8080/healthz`


## Конфигурация (флаги и ENV)

| Флаг/ENV                                | Описание                                         | По умолчанию |
| --------------------------------------- | ------------------------------------------------ | ------------ |
| `--yandex.cloud` / `YC_CLOUD_ID`        | Cloud ID, по которому собирать данные            | (обязателен) |
| `--yandex.token-file` / `YC_TOKEN_FILE` | Путь к файлу с IAM-токеном                       | (обязателен) |
| `--web.listen-address`                  | Адрес для HTTP-сервера                           | `:8080`      |
| `--web.telemetry-path`                  | Путь для метрик                                  | `/metrics`   |
| `--web.max-requests`                    | Лимит параллельных запросов Prometheus           | `40`         |
| `--web.disable-exporter-metrics`        | Прятать служебные метрики exporter’а             | `false`      |
| `--log.level`                           | Уровень логов (`debug`, `info`, `warn`, `error`) | `info`       |
| `--log.format`                          | Формат логов (`json`, `logfmt`)                  | `logfmt`     |


## Как устроен экспортер

Экспортер разделён на два слоя:

1. internal/yandexapi/
Этот пакет знает, как сходить в API Яндекса, получить JSON и превратить его в Go-структуры.

Здесь находится client.go — универсальный клиент, который:

- читает токен из файла (getToken)

- умеет делать GET-запрос (apiGet)

- хранит интерфейс Client с методами для конкретных ресурсов.

Все конкретные запросы (ListInstancesByCloud, ListQuotaLimits, ListFolders) реализуются в отдельных файлах (instances.go, quotas.go, folders.go).

При добавлении нового ресурса (disks.go, databases.go) мы не копировали сетевую логику, а использовали общий client.go.

2. collector/
Этот пакет отвечает за Prometheus-метрики.

collector.go — общий "агрегатор" всех коллекторов (он просто вызывает их по очереди).

Каждый отдельный ресурс (например, yandex_instances.go) реализует свой Collector. 

Collector:

- вызывает методы из yandexapi

- конвертирует данные в метрики Prometheus (prometheus.NewDesc, MustNewConstMetric)

- регистрируется в основном бинарнике.


## Алгоритм для любого нового ресурса

1. В internal/yandexapi/

- создать файл <resource>.go

- описать структуру ответа API и метод List<Resource>ByCloud.

2. В collector/

- создать файл yandex_<resource>.go

- описать Collector (аналогично InstancesCollector).

3. В cmd/yandex-exporter.go

- зарегистрировать новый Collector.