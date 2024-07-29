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
- Number of burned QUs
- Rich list

## Architecture

The service is made up of three parts:
- `MongoDB`
- `Processor`
- `API`

### The Processor
The processor is responsible for calculating and saving the information to the database.

It has two modes of operation: `service` and `spectrum_parser`.
The `service` mode will continuously scrape, calculate and save data.
The `spectrum_parser` mode is used to calculate and save spectrum related data in the database.

> Please refer to `./setupSpectrumData.sh` for an example on how to use `spectrum_parser` mode.

#### Configuration:
```bash
--app-mode/$QUBIC_STATS_PROCESSOR_APP_MODE                                            <string>    (default: service)

--spectrum-parser-spectrum-size/$QUBIC_STATS_PROCESSOR_SPECTRUM_PARSER_SPECTRUM_SIZE  <string>    (default: 16777216)
--spectrum-parser-spectrum-file/$QUBIC_STATS_PROCESSOR_SPECTRUM_PARSER_SPECTRUM_FILE  <string>    (default: ./latest.118)
--spectrum-parser-output-mode/$QUBIC_STATS_PROCESSOR_SPECTRUM_PARSER_OUTPUT_MODE      <string>    (default: db)
--spectrum-parser-output-file/$QUBIC_STATS_PROCESSOR_SPECTRUM_PARSER_OUTPUT_FILE      <string>    (default: spectrumData.json)

--service-archiver-grpc-address/$QUBIC_STATS_PROCESSOR_SERVICE_ARCHIVER_GRPC_ADDRESS  <string>    (default: localhost:8001)
--service-coin-gecko-token/$QUBIC_STATS_PROCESSOR_SERVICE_COIN_GECKO_TOKEN            <string>    
--service-data-scrape-interval/$QUBIC_STATS_PROCESSOR_SERVICE_DATA_SCRAPE_INTERVAL    <duration>  (default: 1m)
--service-data-scrape-timeout/$QUBIC_STATS_PROCESSOR_SERVICE_DATA_SCRAPE_TIMEOUT      <duration>  (default: 5s)

--mongo-username/$QUBIC_STATS_PROCESSOR_MONGO_USERNAME                                <string>    (default: user)
--mongo-password/$QUBIC_STATS_PROCESSOR_MONGO_PASSWORD                                <string>    (default: pass)
--mongo-hostname/$QUBIC_STATS_PROCESSOR_MONGO_HOSTNAME                                <string>    (default: localhost)
--mongo-port/$QUBIC_STATS_PROCESSOR_MONGO_PORT                                        <string>    (default: 27017)
--mongo-options/$QUBIC_STATS_PROCESSOR_MONGO_OPTIONS                                  <string>    
--mongo-database/$QUBIC_STATS_PROCESSOR_MONGO_DATABASE                                <string>    (default: qubic_frontend)
--mongo-spectrum-collection/$QUBIC_STATS_PROCESSOR_MONGO_SPECTRUM_COLLECTION          <string>    (default: spectrum_data)
--mongo-data-collection/$QUBIC_STATS_PROCESSOR_MONGO_DATA_COLLECTION                  <string>    (default: general_data)
--mongo-rich-list-collection/$QUBIC_STATS_PROCESSOR_MONGO_RICH_LIST_COLLECTION        <string>    (default: rich_list)

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
  
  --mongo-username/$QUBIC_STATS_API_MONGO_USERNAME                                                <string>    (default: user)
  --mongo-password/$QUBIC_STATS_API_MONGO_PASSWORD                                                <string>    (default: pass)
  --mongo-hostname/$QUBIC_STATS_API_MONGO_HOSTNAME                                                <string>    (default: localhost)
  --mongo-port/$QUBIC_STATS_API_MONGO_PORT                                                        <string>    (default: 27017)
  --mongo-options/$QUBIC_STATS_API_MONGO_OPTIONS                                                  <string>    
  --mongo-database/$QUBIC_STATS_API_MONGO_DATABASE                                                <string>    (default: qubic_frontend)
  --mongo-spectrum-collection/$QUBIC_STATS_API_MONGO_SPECTRUM_COLLECTION                          <string>    (default: spectrum_data)
  --mongo-data-collection/$QUBIC_STATS_API_MONGO_DATA_COLLECTION                                  <string>    (default: general_data)
  --mongo-rich-list-collection/$QUBIC_STATS_API_MONGO_RICH_LIST_COLLECTION                        <string>    (default: rich_list)
  
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
