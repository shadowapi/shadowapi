FROM arigaio/atlas:1.1.0

COPY db/schema.sql /db/schema.sql
COPY db/tg.sql /db/tg.sql
COPY devops/docker/migrate/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
