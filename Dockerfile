FROM golang:1.27.3-alpine3.23 AS builder
RUN apk --update add build-base nodejs npm

WORKDIR /src/app
# Adding go.mod and go.sum and downloading dependencies first
# This is done to leverage Docker layer caching
ADD go.* .
RUN go mod download

ADD package.json package-lock.json ./
RUN npm ci

ADD . .

RUN npm run build

# Building the migrate command with necessary tags
RUN go build -tags osusergo,netgo -o bin/migrate ./cmd/migrate

# Building the app with necessary tags
RUN go build -tags osusergo,netgo -o bin/app ./cmd/app

FROM alpine
RUN apk add --no-cache tzdata ca-certificates sqlite
ENV GO_ENV=production

WORKDIR /bin/

# Copying binaries to /bin from the builder stage
COPY --from=builder /src/app/bin/app .
COPY --from=builder /src/app/bin/migrate .

# Specifying the shell to use
SHELL ["/bin/ash", "-c"]
CMD migrate && app
