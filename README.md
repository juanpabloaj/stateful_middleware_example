# stateful_middleware_example

Run

    go run main.go

Send a request

    curl 0.0.0.0:8080

Change the middleware behavior to verbose

    curl -d '{"option": "verbose"}' 0.0.0.0:8080/config

Change the middleware behavior to no verbose

    curl -d '{"option": "no_verbose"}' 0.0.0.0:8080/config
