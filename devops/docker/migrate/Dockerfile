FROM arigaio/atlas:1.1.0 AS atlas-bin

FROM alpine:3.21

COPY --from=atlas-bin /atlas /usr/local/bin/atlas

COPY db/schema.sql /db/schema.sql
COPY db/tg.sql /db/tg.sql
COPY --chmod=755 devops/docker/migrate/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
