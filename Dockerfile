FROM golang:1.17.7-alpine3.15 as base

ENV APP /go/src
WORKDIR $APP

RUN apk add -u build-base

COPY go.mod $APP
# COPY go.sum $APP

RUN go mod download

COPY . $APP

FROM base as build

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/app ./main.go

FROM alpine as release

COPY --from=build /go/bin/app /

CMD [ "./app" ]