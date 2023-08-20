# Release History

## v0.5.3 (2023-08-20)

- feat: replace exec with myexec, add several exec helpers
- feat: create python3 venv with funppy if python3 not specified
- fix: create log file directory if not exists in InitLogger

## v0.5.2 (2023-08-10)

- feat: add Init option `WithDebugLogger(debug bool)` to configure whether to print debug level logs in plugin process
- feat: add Init option `WithLogFile(logFile string)` to specify log file path
- feat: add Init option `WithDisableTime(disable bool)` to configure whether disable log time
- refactor: merge shared utils `CallFunc` to fungo package
- refactor: replace zerolog with hclog
- refactor: optimize log printing for plugin
- fix: ensure using grpc for hashicorp python plugin
