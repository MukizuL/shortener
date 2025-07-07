# ========== Phase 1: metadata stage ==========
FROM golang:1.24 AS metadata-stage

WORKDIR /meta
COPY . .

# Git must be available; it is in golang:1.x base image
SHELL ["/bin/bash", "-c"]
RUN \
  BUILD_COMMIT=$(git rev-parse --short HEAD) && \
  BUILD_DATE=$(date +'%Y/%m/%d %H:%M:%S') && \
  echo "COMMIT=$BUILD_COMMIT" >> build.env && \
  echo "DATE=\"$BUILD_DATE\"" >> build.env && \
  echo "VERSION=v0.6.0" >> build.env

# ========== Phase 2: build Go binary ==========
FROM golang:1.24 AS build-stage

WORKDIR /app

COPY --from=metadata-stage /meta/build.env .
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY internal/migration/migrations ./migrations
COPY ./docs ./docs

WORKDIR /app/cmd/shortener

# Load values from build.env
SHELL ["/bin/bash", "-c"]
RUN source /app/cmd/shortener/../../build.env && \
    CGO_ENABLED=0 GOOS=linux \
    go build -ldflags "-X main.buildVersion=$VERSION -X 'main.buildDate=$DATE' -X main.buildCommit=$COMMIT" -o /api

# ========== Phase 3: minimal runtime image ==========
FROM scratch AS run-stage

WORKDIR /app
COPY --from=metadata-stage /meta/build.env .
COPY --from=build-stage /api /api
COPY ./tls ./tls

CMD ["/api"]
