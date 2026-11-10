FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=none
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/mimic ./cmd/mimic

# Mimic only parses source, so the runtime image needs no Go toolchain.
FROM alpine:3.20
RUN adduser -D -u 10001 mimic
COPY --from=build /out/mimic /usr/local/bin/mimic
COPY --from=build /src/deploy/sample /srv/sample
USER mimic
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mimic"]
CMD ["serve", "-dir", "/srv/sample", "-listen", "0.0.0.0:8080"]
