ARG GO_IMAGE=golang:1.25-alpine3.22
ARG NODE_IMAGE=node:23-alpine3.22
ARG RUNTIME_IMAGE=alpine:3.22

# Vue front

FROM ${NODE_IMAGE} AS front-build

WORKDIR /src/web

COPY web/package*.json ./
RUN npm ci

COPY web/ ./      
RUN npm run build

# Go back

FROM ${GO_IMAGE} AS back-build

ARG GOOSE_VERSION=v3.27.2

RUN apk add --no-cache  build-base

WORKDIR /src

COPY go.mod go.sum /

RUN go mod download

COPY . .

COPY --from=front-build /src/web/dist ./web/dist

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -o /out/librorum ./cmd

RUN CGO_ENABLED=1 GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

# Runtime image

FROM ${RUNTIME_IMAGE}

WORKDIR /app

COPY --from=back-build --chown=10002:10002 /out/librorum /app/librorum
COPY --from=back-build --chown=10002:10002 /out/goose /app/goose
COPY --from=front-build --chown=10002:10002 /src/web/dist /app/web/dist

COPY --chown=10002:10002 internal/platform/storage/migrations /app/migrations

#psql

USER 10002:10002

EXPOSE 4545

ENTRYPOINT [ "/app/librorum" ]
