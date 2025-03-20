FROM postgres:14

USER root

RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    postgresql-server-dev-14 \
    && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/pgq/pgq.git /tmp/pgq && \
    cd /tmp/pgq && \
    make PG_CONFIG=/usr/bin/pg_config && \
    make install PG_CONFIG=/usr/bin/pg_config && \
    rm -rf /tmp/pgq

USER postgres
