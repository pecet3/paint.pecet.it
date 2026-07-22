FROM node:latest AS frontend-builder
WORKDIR /src
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ . 
RUN npm run build


FROM golang:latest as backend-builder
WORKDIR /src
COPY backend/go.mod backend/go.sum* ./
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app cmd/main.go


FROM alpine:latest
WORKDIR /app
COPY --from=backend-builder /src/app .
COPY --from=frontend-builder /src/dist ./dist
RUN chmod +x app
EXPOSE 8080
ENTRYPOINT ["./app"]