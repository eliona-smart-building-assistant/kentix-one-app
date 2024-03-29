# Eliona app to access KentixONE devices
This [Eliona app for KentixONE](https://github.com/eliona-smart-building-assistant/kentix-one-app) connects the [KentixONE devices](https://kentix.com/en/kentixone/) to an [Eliona](https://www.eliona.io/) environment.

This app collects data from KentixONE devices such as AccessManager, AlarmManager and MultiSensor and passes their data to Eliona. Each device corresponds to an asset in an Eliona project.

NOTE: This app is passing numeric data to Eliona as strings. That was done as a way to overcome inconsistent data formats that the Kentix API provides. This causes that aggregation is not working for the Kentix data. Before deployment, we should change this and work around the inconsistent data formats provided.

## Configuration

The app needs environment variables and database tables for configuration. To edit the database tables the app provides an own API access.

### Kentix devices setup ###

The KentixONE devices should be interconnected. One device (i.e. AlarmManager) should be in "Manager" mode and other devices should be in "Satellite" mode and connected to the Manager (which is done in the Kentix device UI configuration).

The user then sets the address of the Manager as one Configuration, and the Manager's Satellites are then discovered automatically by the app.

The app can have multiple Configurations - for multiple Managers.

For communication with Kentix devices, the app needs an API key. The key is set in the Kentix device UI configuration. The app then passes it as a bearer token in Authorization header.

### Registration in Eliona ###

To start and initialize an app in an Eliona environment, the app has to registered in Eliona. For this, an entry in the database table `public.eliona_app` is necessary.

The registration could be done using the reset script.

### Environment variables ###

- `CONNECTION_STRING`: configures the [Eliona database](https://github.com/eliona-smart-building-assistant/go-eliona/tree/main/db). Otherwise, the app can't be initialized and started (e.g. `postgres://user:pass@localhost:5432/iot`).

- `INIT_CONNECTION_STRING`: configures the [Eliona database](https://github.com/eliona-smart-building-assistant/go-eliona/tree/main/db) for app initialization like creating schema and tables (e.g. `postgres://user:pass@localhost:5432/iot`). Default is content of `CONNECTION_STRING`.

- `API_ENDPOINT`:  configures the endpoint to access the [Eliona API v2](https://github.com/eliona-smart-building-assistant/eliona-api). Otherwise, the app can't be initialized and started. (e.g. `http://api-v2:3000/v2`)

- `API_TOKEN`: defines the secret to authenticate the app and access the API.

- `API_SERVER_PORT`(optional): defines the port the API server listens on. The default value is `3000`.

- `LOG_LEVEL`(optional): defines the minimum level that should be [logged](https://github.com/eliona-smart-building-assistant/go-utils/blob/main/log/README.md). Not defined the default level is `info`.

### Database tables ###

The app requires configuration data that remains in the database. To do this, the app creates its own database schema `kentix_one` during initialization. To modify and handle the configuration data the app provides an API access. Have a look at the [API specification](https://eliona-smart-building-assistant.github.io/open-api-docs/?https://raw.githubusercontent.com/eliona-smart-building-assistant/kentix-one-app/develop/openapi.yaml) how the configuration tables should be used.

- `kentix-one.configuration`: Configurations for individual KentixONE devices. Editable by API.

- `kentix-one.device`: Specific devices, one for each project and configuration. One device corresponds to one asset in Eliona.

There is 1:N relationship between configuration and device (i.e. one Configuration could be in multiple projects and each would have it's own device).

**Generation**: to generate access method to database see Generation section below.


## References

### KentixONE App API ###

The KentixONE app provides its own API to access configuration data and other functions. The full description of the API is defined in the `openapi.yaml` OpenAPI definition file.

- [API Reference](https://eliona-smart-building-assistant.github.io/open-api-docs/?https://raw.githubusercontent.com/eliona-smart-building-assistant/kentix-one-app/develop/openapi.yaml) shows details of the API

**Generation**: to generate api server stub see Generation section below.


### Eliona assets ###

This app creates Eliona asset types and attribute sets during initialization.

The data is written for each KentixONE device, structured into different subtypes of Elinoa assets. The following subtypes are defined:

- `Info`: Static data which specifies a KentixONE device like address and firmware info.
- `Input`: Current values reported by KentixONE sensors (i.e. MultiSensor readings).

This app also allows the KentixONE devices to be controlled from Eliona environment, by `Output` subtypes.

### Continuous asset creation ###

Assets for all devices connected to the configured "Manager" device are created automatically when the configuration is added.

To select which assets to create, a filter could be specified in config. The schema of the filter is defined in the `openapi.yaml` file. Please note that regex special characters have to be double-escaped with `\\` to avoid getting interpreted.

Possible filter parameters are defined in following places:

- `kentix/kentix.go:DeviceInfo` for devices
- `kentix/kentix.go:Doorlock` for doorlocks

## Tools

### Generate API server stub ###

For the API server the [OpenAPI Generator](https://openapi-generator.tech/docs/generators/openapi-yaml) for go-server is used to generate a server stub. The easiest way to generate the server files is to use one of the predefined generation script which use the OpenAPI Generator Docker image.

```
.\generate-api-server.cmd # Windows
./generate-api-server.sh # Linux
```

### Generate Database access ###

For the database access [SQLBoiler](https://github.com/volatiletech/sqlboiler) is used. The easiest way to generate the database files is to use one of the predefined generation script which use the SQLBoiler implementation. Please note that the database connection in the `sqlboiler.toml` file have to be configured.

```
.\generate-db.cmd # Windows
./generate-db.sh # Linux
```

### Mock the KentixONE devices ###
`kentix-one-mock` folder contains mock endpoints implementation:
- `access-manager`: `localhost:3031`
- `alarm-manager`: `localhost:3032`
- `multi-sensor`: `localhost:3033`

The [Kentix API documentation](https://kentix.com/transfer/smartapi/alarmmanager) roughly corresponds to the device APIs, but there are subtle differences (like having different data type for doorlock battery level). Kentix promised to document the KentixONE API in 2023.
