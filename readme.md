## Run in docker

    1.  setup docker
    2.  docker build -t btcdata .
    3.  docker run -e DB_HOST=192.168.99.100 --rm -t btcdata (DB_HOST should be set to the hosts IP)


    - set environment variables: -e VAR_TE=99

## Run in raw
    1.  go run main.go

## Local Testing Setup

#### Postgres

postgresql.conf

    listen_addresses = '*'

pg_hba.conf

    # TYPE  DATABASE        USER            ADDRESS                 METHOD

    # "local" is for Unix domain socket connections only
    local   all             all                                     trust
    # IPv4 local connections:
    host    all             all             127.0.0.1/32            trust
    # IPv6 local connections:
    host    all             all             ::1/128                 trust
    host     all             all             0.0.0.0/0                 trust
    # Allow replication connections from localhost, by a user with the
    # replication privilege.
    #local   replication     sogasg                                trust
    #host    replication     sogasg        127.0.0.1/32            trust
    #host    replication     sogasg        ::1/128                 trust

With problems: RESTART COMPUTER :P

## ToDo

-   make sure only one instance is running at the time (make sure ticks do not duplucate)