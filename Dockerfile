FROM golang:1.18.2-alpine3.15 as server-builder

WORKDIR /app

COPY go.mod go.sum main.go cli.go ./

RUN go build -o bin/jpegme .

FROM node:16-alpine3.15 as frontend-builder

WORKDIR /app

COPY package.json package-lock.json ./

RUN npm install

COPY main.js webpack.config.js .babelrc ./

RUN npx webpack

FROM alpine:3.15.4
ARG PORT=5000
ENV PORT=$PORT
EXPOSE $PORT

WORKDIR /app

COPY --from=server-builder /app/bin/jpegme /app/bin/jpegme
COPY --from=frontend-builder /app/static/ /app/static/
COPY static/index.html /app/static/

CMD ["/app/bin/jpegme"]
