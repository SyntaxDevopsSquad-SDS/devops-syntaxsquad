FROM golang:1.25 AS build

WORKDIR /src/implementations/go

COPY implementations/go/go.mod implementations/go/go.sum ./
RUN go mod download

COPY implementations/go/backend ./backend
COPY implementations/go/migrations ./migrations
COPY implementations/go/templates ./templates
COPY implementations/go/static ./static
COPY implementations/go/schema.sql ./schema.sql

RUN cd backend && go build -o /out/whoknows .

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app/backend

COPY --from=build /out/whoknows /app/backend/whoknows
COPY --from=build /src/implementations/go/migrations /app/migrations
COPY --from=build /src/implementations/go/templates /app/templates
COPY --from=build /src/implementations/go/static /app/static
COPY --from=build /src/implementations/go/schema.sql /app/schema.sql
COPY implementations/go/backend/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh


EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
