# Build the manager binary
FROM golang:1.17

ARG CCLOUD_EMAIL
ARG CCLOUD_PASSWORD

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the Confluent requisites
COPY confluent/cmd/confluent /bin/confluent
COPY confluent/script/.netrc /root

RUN  sed -i -r "s/ccloudlogin/${CCLOUD_EMAIL}/g" /root/.netrc
RUN  sed -i -r "s/ccloudpassword/${CCLOUD_PASSWORD}/g" /root/.netrc

RUN go mod download

# Copy the go source
# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY services/ services/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o runner main.go

USER root

RUN mkdir /manager && \
    mv /workspace/runner /manager/. && \
    chmod -R 777 /manager

RUN ls -lha /manager

CMD ["/manager/runner","-D","FOREGROUND"]