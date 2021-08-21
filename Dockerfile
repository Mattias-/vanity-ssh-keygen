FROM golang:1.17 as builder

WORKDIR /src
COPY ./ /src
RUN go mod tidy
RUN CGO_ENABLED=0 go build -v -trimpath ./cmd/vanity-ssh-keygen

FROM scratch
COPY --from=builder /src/vanity-ssh-keygen /vanity-ssh-keygen
ENTRYPOINT ["/vanity-ssh-keygen"]
