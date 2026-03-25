module go-ser

go 1.25.0

replace github.com/acoinfo/go-isatty => github.com/acoinfo/go-isatty v0.0.20-sylixos.1

replace github.com/acoinfo/sys => github.com/acoinfo/sys v0.35.0-sylixos.5

replace github.com/acoinfo/logrus => github.com/acoinfo/logrus v1.9.3-sylixos.1

replace github.com/sirupsen/logrus => github.com/acoinfo/logrus v1.9.3-sylixos.1

replace golang.org/x/sys => github.com/acoinfo/sys v0.35.0-sylixos.5

replace github.com/mattn/go-isatty => github.com/acoinfo/go-isatty v0.0.20-sylixos.1

replace go.etcd.io/bbolt => github.com/acoinfo/bbolt v1.4.2-sylixos.1

require (
	github.com/acoinfo/vsoa v1.1.10
	github.com/gorilla/websocket v1.5.3
	github.com/julienschmidt/httprouter v1.3.0
	github.com/ncruces/go-sqlite3 v0.30.4
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/tetratelabs/wazero v1.11.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
)
