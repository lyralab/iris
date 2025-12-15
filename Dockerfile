FROM golang:1.25-alpine as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0;GOEXPERIMENT=greenteagc go build -o /go/bin/app ./cmd/server/main.go


FROM node:20-alpine as node_build
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ .
RUN npm run build

# Now copy it into our base image.
FROM chainguard/wolfi-base:latest
COPY --from=build /go/src/app/migrations /migrations
COPY --from=build /go/src/app/assets /assets
COPY --from=node_build /app/web/build /web/build
COPY --from=build /go/bin/app /
CMD ["/app"]