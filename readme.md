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

-   move topics to europe
-   external config
-   make sure only one instance is running at the time (make sure ticks do not duplucate)