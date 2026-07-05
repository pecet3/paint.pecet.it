FROM alpine:latest

WORKDIR /app

COPY backend/app .
COPY frontend-paint.pecet.it/dist ./dist

RUN chmod +x app

EXPOSE 8080

ENTRYPOINT ["./app"]