FROM harbor.jibit.cloud/library/chainguard/go:latest-dev AS build

WORKDIR /app
COPY . .

ENV GOPROXY=https://nexus.jibit.cloud/repository/gobit/
RUN go mod download
RUN CGO_ENABLED=0 go build -o /usr/bin/app ./cmd/server/main.go


FROM harbor.jibit.cloud/library/chainguard/node:latest-dev AS node_build
WORKDIR /app
COPY --chown=node:node web/package*.json ./
RUN npm install
COPY --chown=node:node web/ .
RUN npm run build

# Now copy it into our base image.
FROM harbor.jibit.cloud/library/chainguard/static:latest
COPY --from=build /app/migrations /migrations
COPY --from=build /app/assets /assets
COPY --from=node_build /app/build /web/build
COPY --from=build /usr/bin/app /app
CMD ["/app"]
