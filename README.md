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
2. Установить зависимости и собрать:

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

### 1. `internal/yandexapi/`
Этот пакет знает, как сходить в API Яндекса, получить JSON и превратить его в Go-структуры.

Здесь находится `client.go` — универсальный клиент, который:

- читает токен из файла (`getToken`)
- делает GET-запрос с пагинацией (`apiPagedGet`)
- хранит интерфейс `Client` с методами для конкретных ресурсов.

Все конкретные запросы (`ListInstancesByCloud`, `ListQuotaLimits`, `ListFolders`) реализуются в отдельных файлах (`instances.go`, `quotas.go`, `folders.go`).

При добавлении нового ресурса (например, `disks.go`, `databases.go`) мы не копируем сетевую логику, а используем общий `client.go`.

---

### 2. `collector/`
Этот пакет отвечает за Prometheus-метрики.

- `collector.go` — общий "агрегатор" всех коллекторов (он просто вызывает их по очереди).
- Каждый отдельный ресурс (например, `yandex_instances.go`) реализует свой `Collector`.

Collector:

- вызывает методы из `yandexapi`
- конвертирует данные в метрики Prometheus (`prometheus.NewDesc`, `MustNewConstMetric`)
- регистрируется в основном бинарнике.


## Алгоритм для добавления любого нового ресурса

### 1. `В internal/yandexapi/`

- создать файл <resource>.go

- описать структуру ответа API и метод List<Resource>ByCloud. 

Пример:

```go
package yandexapi

import (
    "context"
    "fmt"
)

type MyResource struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func (c *client) ListMyResourcesByCloud(cloudID string) ([]MyResource, error) {
    token, err := c.getToken()
    if err != nil {
        return nil, err
    }

    var all []MyResource
    err = apiPagedGet(
        context.Background(),
        c.httpCli,
        token,
        func(pageToken string) string {
            if pageToken == "" {
                return fmt.Sprintf("https://example.api.cloud.yandex.net/v1/resources?cloudId=%s&pageSize=1000", cloudID)
            }
            return fmt.Sprintf("https://example.api.cloud.yandex.net/v1/resources?cloudId=%s&pageSize=1000&pageToken=%s", cloudID, pageToken)
        },
        "resources", // ключ массива в JSON
        &all,
    )
    if err != nil {
        return nil, fmt.Errorf("list my resources failed: %w", err)
    }

    return all, nil
}
```

### 2. `В collector/`

- создать файл yandex_<resource>.go

- описать Collector (аналогично InstancesCollector).

Пример:

```go
package collector

import (
    "log/slog"
    "yandex_exporter/internal/yandexapi"

    "github.com/prometheus/client_golang/prometheus"
)

type MyResourceCollector struct {
    api     yandexapi.Client
    cloudID string
    info    *prometheus.Desc
}

func NewMyResourceCollector(api yandexapi.Client, cloudID string) *MyResourceCollector {
    return &MyResourceCollector{
        api:     api,
        cloudID: cloudID,
        info: prometheus.NewDesc(
            prometheus.BuildFQName(namespace, "myresource", "info"),
            "Yandex MyResource information",
            []string{"cloud", "id", "name"},
            nil,
        ),
    }
}

func (c *MyResourceCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.info
}

func (c *MyResourceCollector) Collect(ch chan<- prometheus.Metric) {
    items, err := c.api.ListMyResourcesByCloud(c.cloudID)
    if err != nil {
        slog.Debug("failed to list my resources", "err", err) // DEBUG, не ломаем scrape
        return
    }

    for _, r := range items {
        ch <- prometheus.MustNewConstMetric(
            c.info,
            prometheus.GaugeValue,
            1,
            c.cloudID, r.ID, r.Name,
        )
    }
}
```

### 3. `В cmd/yandex-exporter.go`

- зарегистрировать новый Collector.

```go
api := yandexapi.NewClient(*ycTokenFile)

instCollector := collector.NewInstancesCollector(api, *ycCloud)
quotaCollector := collector.NewQuotaCollector(api, *ycCloud)
myResCollector := collector.NewMyResourceCollector(api, *ycCloud)

reg := prometheus.NewRegistry()
reg.MustRegister(instCollector, quotaCollector, myResCollector)
```