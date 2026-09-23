# The Qubic Stats Service

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

It has three modes of operation: `service`, `spectrum_parser` and `backfill_epoch_stats`.
The `service` mode will continuously scrape, calculate and save data.
The `spectrum_parser` mode is used to calculate and save spectrum related data in the database.
The `backfill_epoch_stats` mode is a one off migration, described below.

> Please refer to `./setupSpectrumData.sh` for an example on how to use `spectrum_parser` mode.

#### Supply history

The `service` mode keeps one record per epoch in the `epoch_stats` collection, holding the
circulating supply at the end of the epoch, the cumulative issuance up to it and when the epoch
closed. Records are keyed by epoch, so writing one is idempotent. The API serves them through
`/v1/stats/supply-history`.

A spectrum file named for epoch `N` is dumped at the transition into `N` and holds the **end state of
epoch `N-1`**, so it is the authoritative measurement of epoch `N-1` and pairs with `(N-1) * 1e12` of
issuance. Records therefore only ever exist for epochs that have closed; the epoch in progress has no
data point until its own spectrum file has been parsed, and is reported separately by the API as
`currentEpoch` and `currentCirculatingSupply`.

Because the measurement carries the epoch it belongs to, it does not matter how long after the
transition the spectrum file turns up. A file that only becomes available once the following epoch
has started is still attributed to the epoch it actually measured.

Every record states where its numbers came from:

| `supplySource`  | Meaning                                                                        |
|-----------------|--------------------------------------------------------------------------------|
| `spectrum`      | Measured from the spectrum file that closed the epoch. Authoritative.           |
| `derived`       | Reconstructed by the backfill from the historical general data. Exact value.    |
| `carried-over`  | Backfill only: no spectrum file could be attributed, so an earlier supply is repeated. |

A `carried-over` epoch shows no burn, because nothing was measured for it. It points at a missed
`spectrum_parser` run rather than at a real network event. Live operation never produces one: an
epoch with no measurement simply gets no record.

`timestampSource` is `tick` when the boundary was read from the first tick carrying tick data in the
following epoch, and `first-observed` when it falls back to when the spectrum file was parsed, which
runs shortly after the transition.

#### Backfilling the supply history

The supply history before this feature existed can be reconstructed from the `general_data`
collection, which kept the burned QUs and the epoch of every scrape. The backfill inverts
`burnedQUs = epoch * 1e12 - circulatingSupply`, joins the spectrum measurements in by timestamp and
writes the epochs that have no record yet. Records that already exist are never modified, so the
command is safe to re-run.

The supply an epoch `A` ran on was measured at the end of `A-1`, so what the general data of epoch
`A` reconstructs is recorded under epoch `A-1`.

```bash
processor/qubic-stats-processor --app-mode=backfill_epoch_stats --mongo-username user --mongo-password pass
```

It prints the reconstructed range and warns about epochs with no general data, epochs whose supply is
carried over, and epochs where the spectrum measurement disagrees with the derived supply. Check that
output before applying any retention to `general_data`: that collection is the only place the
historical supply exists.

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
--mongo-spectrum-collection/$QUBIC_STATS_PROCESSOR_MONGO_SPECTRUM_COLLECTION                    <string>    (default: spectrum_data)
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
--service-spectrum-data-update-interval/$QUBIC_STATS_API_SERVICE_SPECTRUM_DATA_UPDATE_INTERVAL  <duration>  (default: 24h)
--service-rich-list-page-size/$QUBIC_STATS_API_SERVICE_RICH_LIST_PAGE_SIZE                      <int>       (default: 100)
--service-cache-update-timeout/$QUBIC_STATS_API_SERVICE_CACHE_UPDATE_TIMEOUT                    <duration>  (default: 30s)
--service-rich-list-limit/$QUBIC_STATS_API_SERVICE_RICH_LIST_LIMIT                              <int>       (default: 10000)
--mongo-username/$QUBIC_STATS_API_MONGO_USERNAME                                                <string>    (default: user)
--mongo-password/$QUBIC_STATS_API_MONGO_PASSWORD                                                <string>    (default: pass)
--mongo-hostname/$QUBIC_STATS_API_MONGO_HOSTNAME                                                <string>    (default: localhost)
--mongo-port/$QUBIC_STATS_API_MONGO_PORT                                                        <string>    (default: 27017)
--mongo-options/$QUBIC_STATS_API_MONGO_OPTIONS                                                  <string>    
--mongo-database/$QUBIC_STATS_API_MONGO_DATABASE                                                <string>    (default: qubic_frontend)
--mongo-spectrum-collection/$QUBIC_STATS_API_MONGO_SPECTRUM_COLLECTION                          <string>    (default: spectrum_data)
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

##### /v1/stats/supply-history
Provides the circulating supply per epoch, one point per completed epoch, ordered by epoch ascending.
Each point is the supply at the end of that epoch. The epoch in progress is reported separately as
`currentEpoch` / `currentCirculatingSupply`.

All parameters are optional: `fromEpoch` and `toEpoch` are inclusive and default to the earliest and
the latest available epoch, `limit` caps the number of points and keeps the most recent ones.

```shell
curl "http://127.0.0.1:8080/v1/stats/supply-history?fromEpoch=179&toEpoch=180"
```

```json
{
  "supplyCap": "1000000000000000",
  "currentEpoch": 180,
  "currentCirculatingSupply": "149832400000000",
  "points": [
    {
      "epoch": 179,
      "circulatingSupply": "149218900000000",
      "totalEmitted": "179000000000000",
      "timestamp": "1753862400",
      "supplySource": "spectrum"
    },
    {
      "epoch": 180,
      "circulatingSupply": "149832400000000",
      "totalEmitted": "180000000000000",
      "timestamp": "1754467200",
      "supplySource": "spectrum"
    }
  ]
}
```

`currentCirculatingSupply` is read from the same place `/v1/latest-stats` reads it, so the two
endpoints cannot disagree. See the processor section for what `supplySource` means.

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
