ARG REGISTRY=cgr.dev
ARG GOPROXY=https://proxy.golang.org,direct
ARG NPM_REGISTRY=https://registry.npmjs.org/

FROM ${REGISTRY}/chainguard/go:latest-dev AS build
ARG GOPROXY

WORKDIR /app
COPY . .

ENV GOPROXY=${GOPROXY}
RUN go mod download
RUN CGO_ENABLED=0 go build -o /usr/bin/app ./cmd/server/main.go


FROM ${REGISTRY}/chainguard/node:latest-dev AS node_build
ARG NPM_REGISTRY
ENV npm_config_registry=${NPM_REGISTRY}
WORKDIR /app
COPY --chown=node:node web/package*.json ./
RUN npm install
COPY --chown=node:node web/ .
RUN npm run build

# Now copy it into our base image.
FROM ${REGISTRY}/chainguard/static:latest
COPY --from=build /app/migrations /migrations
COPY --from=build /app/assets /assets
COPY --from=node_build /app/build /web/build
COPY --from=build /usr/bin/app /app
CMD ["/app"]
