FROM arigaio/atlas:1.1.0

COPY db/schema.sql /db/schema.sql
COPY db/tg.sql /db/tg.sql
COPY --chmod=755 devops/docker/migrate/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
