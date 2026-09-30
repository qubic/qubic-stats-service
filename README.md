# The Qubic Stats Service

> [!WARNING]
> This version is not compatible with the data written by v0.7.0 and earlier. The circulating supply,
> the rich list and the latest stats are stored differently, so the spectrum files have to be parsed
> again (see [Supply history](#supply-history)) before the service and the API are started. The
> `spectrum_data` and `rich_list_<epoch>` collections and the old `general_data` documents are no
> longer used.

The stats service's purpose is to calculate, save and expose general data related to Qubic.

Currently, the service stores the following data:
- Circulating supply
- Nr. of active addresses 
- Coin price in USD
- Market Cap
- Current Epoch
- Current Tick
- Tick count of current Epoch
- Empty tick count of current Epoch
- Epoch tick quality (ratio between total ticks and non-empty ticks)
- Tick count of the last 10.000 ticks
- Empty tick count of the last 10.000 ticks
- Tick quality of the last 10.000 ticks
- Number of burned QUs
- Rich list
- Circulating supply per epoch (supply history)

## Architecture

The service is made up of three parts:
- `MongoDB`
- `Processor`
- `API`

### The Processor
The processor is responsible for calculating and saving the information to the database.

It has two modes of operation: `service` and `spectrum_parser`.
The `service` mode will continuously scrape, calculate and save data. It keeps a single document in
the `general_data` collection, which every scrape overwrites.
The `spectrum_parser` mode is used to calculate and save spectrum related data in the database.

> Please refer to `./setupSpectrumData.sh` for an example on how to use `spectrum_parser` mode.

#### Supply history

The `spectrum_parser` mode keeps one record per completed epoch in the `epoch_stats` collection. A
record holds the circulating supply and the active addresses at the end of the epoch, the cumulative
issuance up to it and when the epoch ended.

A spectrum file named for epoch `N` is dumped at the transition into `N` and holds the **end state of
epoch `N-1`**, so it produces the record of epoch `N-1`:

| Field               | Value                                                                  |
|---------------------|------------------------------------------------------------------------|
| `circulatingSupply` | Sum of all balances in the file.                                       |
| `activeAddresses`   | Number of addresses in the file.                                       |
| `totalEmitted`      | `(N-1) * 1e12`, as 1T QU is emitted at the end of every epoch.          |
| `epochEndTimestamp` | Wednesday 12:00 UTC at which epoch `N-1` ended, calculated from `N`.    |

A record depends on nothing but the file, so parsing a file again always writes the same record.
The epoch in progress has no record until it has ended and its spectrum file has been parsed.

The most recent record is also the source of the circulating supply, the active addresses and the
burned QUs (`totalEmitted - circulatingSupply`) that `/v1/latest-stats` reports, and of the rich
list's epoch. The service refuses to scrape until at least one record exists.

The rich list is kept in a single collection. Parsing a file replaces it, but only if the file is at
least as recent as the latest record, so older files can be parsed at any time to fill in the
history without touching the current rich list. The new rich list is written to a staging
collection and renamed over the live one, so the API never sees a partial list.

Filling in the history is a matter of parsing every available spectrum file, in any order:

```bash
for file in spectrum.*; do
  processor/qubic-stats-processor --mongo-username user --mongo-password pass --app-mode=spectrum_parser --spectrum-parser-spectrum-file="$file"
done
```

#### Configuration:
```bash
--app-mode/$QUBIC_STATS_PROCESSOR_APP_MODE                                                      <string>    (default: service)
--spectrum-parser-spectrum-file/$QUBIC_STATS_PROCESSOR_SPECTRUM_PARSER_SPECTRUM_FILE            <string>    (default: ./latest.118)
--spectrum-parser-output-mode/$QUBIC_STATS_PROCESSOR_SPECTRUM_PARSER_OUTPUT_MODE                <string>    (default: db)
--spectrum-parser-output-file/$QUBIC_STATS_PROCESSOR_SPECTRUM_PARSER_OUTPUT_FILE                <string>    (default: spectrumData.json)
--service-query-service-grpc-address/$QUBIC_STATS_PROCESSOR_SERVICE_QUERY_SERVICE_GRPC_ADDRESS  <string>    (default: localhost:8001)
--service-live-service-grpc-address/$QUBIC_STATS_PROCESSOR_SERVICE_LIVE_SERVICE_GRPC_ADDRESS    <string>    (default: localhost:8002)
--service-coin-gecko-token/$QUBIC_STATS_PROCESSOR_SERVICE_COIN_GECKO_TOKEN                      <string>    
--service-data-scrape-interval/$QUBIC_STATS_PROCESSOR_SERVICE_DATA_SCRAPE_INTERVAL              <duration>  (default: 1m)
--service-data-scrape-timeout/$QUBIC_STATS_PROCESSOR_SERVICE_DATA_SCRAPE_TIMEOUT                <duration>  (default: 15s)
--mongo-username/$QUBIC_STATS_PROCESSOR_MONGO_USERNAME                                          <string>    (default: user)
--mongo-password/$QUBIC_STATS_PROCESSOR_MONGO_PASSWORD                                          <string>    (default: pass)
--mongo-hostname/$QUBIC_STATS_PROCESSOR_MONGO_HOSTNAME                                          <string>    (default: localhost)
--mongo-port/$QUBIC_STATS_PROCESSOR_MONGO_PORT                                                  <string>    (default: 27017)
--mongo-options/$QUBIC_STATS_PROCESSOR_MONGO_OPTIONS                                            <string>    
--mongo-database/$QUBIC_STATS_PROCESSOR_MONGO_DATABASE                                          <string>    (default: qubic_frontend)
--mongo-data-collection/$QUBIC_STATS_PROCESSOR_MONGO_DATA_COLLECTION                            <string>    (default: general_data)
--mongo-rich-list-collection/$QUBIC_STATS_PROCESSOR_MONGO_RICH_LIST_COLLECTION                  <string>    (default: rich_list)
--mongo-epoch-stats-collection/$QUBIC_STATS_PROCESSOR_MONGO_EPOCH_STATS_COLLECTION              <string>    (default: epoch_stats)

--help/-h
```



### The API
The API is responsible for exposing the stored information.


#### Configuration
```bash
--service-http-address/$QUBIC_STATS_API_SERVICE_HTTP_ADDRESS                                    <string>    (default: 0.0.0.0:8080)
--service-grpc-address/$QUBIC_STATS_API_SERVICE_GRPC_ADDRESS                                    <string>    (default: 0.0.0.0:8081)
--service-cache-validity-duration/$QUBIC_STATS_API_SERVICE_CACHE_VALIDITY_DURATION              <duration>  (default: 10s)
--service-rich-list-page-size/$QUBIC_STATS_API_SERVICE_RICH_LIST_PAGE_SIZE                      <int>       (default: 100)
--service-cache-update-timeout/$QUBIC_STATS_API_SERVICE_CACHE_UPDATE_TIMEOUT                    <duration>  (default: 30s)
--service-rich-list-limit/$QUBIC_STATS_API_SERVICE_RICH_LIST_LIMIT                              <int>       (default: 10000)
--mongo-username/$QUBIC_STATS_API_MONGO_USERNAME                                                <string>    (default: user)
--mongo-password/$QUBIC_STATS_API_MONGO_PASSWORD                                                <string>    (default: pass)
--mongo-hostname/$QUBIC_STATS_API_MONGO_HOSTNAME                                                <string>    (default: localhost)
--mongo-port/$QUBIC_STATS_API_MONGO_PORT                                                        <string>    (default: 27017)
--mongo-options/$QUBIC_STATS_API_MONGO_OPTIONS                                                  <string>    
--mongo-database/$QUBIC_STATS_API_MONGO_DATABASE                                                <string>    (default: qubic_frontend)
--mongo-data-collection/$QUBIC_STATS_API_MONGO_DATA_COLLECTION                                  <string>    (default: general_data)
--mongo-rich-list-collection/$QUBIC_STATS_API_MONGO_RICH_LIST_COLLECTION                        <string>    (default: rich_list)
--mongo-epoch-stats-collection/$QUBIC_STATS_API_MONGO_EPOCH_STATS_COLLECTION                    <string>    (default: epoch_stats)
--mongo-timeout/$QUBIC_STATS_API_MONGO_TIMEOUT                                                  <duration>  (default: 15s)
--pool-node-fetcher-url/$QUBIC_STATS_API_POOL_NODE_FETCHER_URL                                  <string>    (default: http://127.0.0.1:8080/status)
--pool-node-fetcher-timeout/$QUBIC_STATS_API_POOL_NODE_FETCHER_TIMEOUT                          <duration>  (default: 2s)
--pool-node-port/$QUBIC_STATS_API_POOL_NODE_PORT                                                <string>    (default: 21841)
--pool-initial-cap/$QUBIC_STATS_API_POOL_INITIAL_CAP                                            <int>       (default: 5)
--pool-max-idle/$QUBIC_STATS_API_POOL_MAX_IDLE                                                  <int>       (default: 20)
--pool-max-cap/$QUBIC_STATS_API_POOL_MAX_CAP                                                    <int>       (default: 30)
--pool-idle-timeout/$QUBIC_STATS_API_POOL_IDLE_TIMEOUT                                          <duration>  (default: 15s)
--asset-service-ttl/$QUBIC_STATS_API_ASSET_SERVICE_TTL                                          <duration>  (default: 10m)

--help/-h
```

#### Endpoints

This is a brief example on the available endpoints and their responses.
For proper API documentation please refer to the [Swagger File](api/protobuff/stats-api.swagger.json).

##### /v1/latest-stats
Provides the latest available information.

```shell
curl http://127.0.0.1:8080/v1/latest-stats
```

```json
{
  "data": {
    "timestamp": "1722259858",
    "circulatingSupply": "106929085187330",
    "activeAddresses": 476802,
    "price": 0.000001898,
    "marketCap": "202951409",
    "epoch": 119,
    "currentTick": 15102892,
    "ticksInCurrentEpoch": 72892,
    "emptyTicksInCurrentEpoch": 1019,
    "epochTickQuality": 98.602036,
    "burnedQus": "12070914812670"
  }
}
```

##### /v1/supply-history
Provides the circulating supply per epoch, one point per completed epoch, ordered by epoch ascending.
Each point is the supply at the end of that epoch, and its `timestamp` is when the epoch ended. The
epoch in progress is reported separately as `currentEpoch` / `currentCirculatingSupply`.

All parameters are optional: `fromEpoch` and `toEpoch` are inclusive and default to the earliest and
the latest available epoch, `limit` caps the number of points and keeps the most recent ones.

```shell
curl "http://127.0.0.1:8080/v1/supply-history?fromEpoch=230&toEpoch=231"
```

```json
{
  "supplyCap": "1000000000000000",
  "currentEpoch": 232,
  "currentCirculatingSupply": "149832400000000",
  "points": [
    {
      "epoch": 230,
      "circulatingSupply": "149218900000000",
      "totalEmitted": "230000000000000",
      "timestamp": "1789560000"
    },
    {
      "epoch": 231,
      "circulatingSupply": "149832400000000",
      "totalEmitted": "231000000000000",
      "timestamp": "1790164800"
    }
  ]
}
```

`currentCirculatingSupply` is the supply of the latest point, read from the same place
`/v1/latest-stats` reads it, so the two endpoints cannot disagree.

##### /v1/epochs/{epoch}/rich-list

```shell
curl http://127.0.0.1:8080/v1/epochs/{epoch}/rich-list
```

```json
{
  "pagination": {
    "totalRecords": 476802,
    "currentPage": 1,
    "totalPages": 4769
  },
  "richList": {
    "entities": [
      {
        "identity": "BYBYFUMBVLPUCANXEXTSKVMGFCJBMTLPPOFVPNSATABMWDGTMFXPLZLBCXJL",
        "balance": "8647845752843"
      },
      {
        "identity": "VFWIEWBYSIMPBDHBXYFJVMLGKCCABZKRYFLQJVZTRBUOYSUHOODPVAHHKXPJ",
        "balance": "3220439015928"
      },
      {
        "identity": "QZYSHUTAJSTEXAYKBOVSOSSQQXHDHZAVNXLJFYOKVCZTJXPBQNDLRODBZXUC",
        "balance": "2130000088435"
      },
      {
        "identity": "TEJNFRXFYVLJBBGCIMOTUZVTJGUCDRMZTVPDXRPWKGQEDWCVHWTAGKBHMEHM",
        "balance": "2027481518128"
      },
      {
        "identity": "VLPRWVPIOMSSDFWOZCMKIYNSTKHBBZANIGXOXQXACCRYTORTANHHTYPFGVHF",
        "balance": "1962644306498"
      },
      {
        "identity": "PSSAMNSRCLRMJAEAMCLOYGMYXUDBZUDLSXXFEXXCNEFKFESORYKLBQIBVFKJ",
        "balance": "1898000000000"
      },
      {
        "identity": "JBPPTAOMVKOTVFBCHHZQIVHWAZRAJYSLDVBKFSCZVAWRTMEVWYWCNEZAAYSH",
        "balance": "1760678088434"
      },
      {
        "identity": "IZTNWDKXSFULQADTOLTMLUPHSCFCXLOJMQOUHPBSRGQZMMXZCJYQFTRDOGRE",
        "balance": "1621432612691"
      },
      .
      .
      .
    ]
  }
}
```
