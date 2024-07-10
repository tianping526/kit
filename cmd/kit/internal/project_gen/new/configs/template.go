package configs

const apolloTemplate = `
{{- /* delete empty line */ -}}
#apollo:
#  endpoint: http://localhost:8080
#  cluster: default
#  secret: faker
`

const serviceTemplate = `
{{- /* delete empty line */ -}}
bootstrap:
#  trace:
#    endpoint: http://127.0.0.1:14268/api/traces
  server:
    http:
      addr: 0.0.0.0:8000
      timeout: 1s
    grpc:
      addr: 0.0.0.0:9000
      timeout: 1s
  data:
    database:
      driver: postgres
      source: postgresql://postgres:example@127.0.0.1:12211/postgres
#      driver: mysql
#      source: root:example@tcp(127.0.0.1:12213)/test?{{- /* delete empty */ -}}
		timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8&interpolateParams=true
      max_open: 100
      max_idle: 10
      conn_max_life_time: 0s
      conn_max_idle_time: 300s
    redis:
      addr: 127.0.0.1:12212
      password:
      db_index: 0
      dial_timeout: 1s
      read_timeout: 0.2s
      write_timeout: 0.2s
#  registry:
#    consul:
#      address: 127.0.0.1:8500
#      scheme: http
  auth:
    key: some-secret-key
#  log:
#    level: INFO # DEBUG, INFO, WARN, ERROR
#    encoding: JSON # JSON, CONSOLE
#    sampling:
#      initial: 100
#      thereafter: 100
#    output_paths:
#      - path: ./log/server.log
#        rotate:
#          max_size: 100
#          max_age: 30
#          max_backups: 3
#          compress: false
`
