FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY momentic-backend .

RUN chmod +x ./momentic-backend

EXPOSE 8080

CMD ["./momentic-backend"]
