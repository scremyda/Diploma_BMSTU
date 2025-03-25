FROM postgres:14

USER root

RUN apt-get update && apt-get install -y \
    postgresql-14-pgq3 \
    pgqd

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod +x /usr/local/bin/entrypoint.sh

USER postgres

COPY pgqd.ini /etc/pgqd.ini

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

CMD ["postgres"]