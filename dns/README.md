# DNS Daemon

## Description

DNS Daemon is a microservice developed in Go, providing DNS management functionality via a REST API. <br />
It allows users to create, update, delete and retrieve DNS information. <br />
Additional utility functions include checking the status of a DNS record and retrieving the gateway IP address. <br />

## Features

- Manage DNS records, including creating, updating and deleting DNS information.
- Retrieve specific DNS record information.
- Check the status of a DNS record.
- Get the Gateway IP address.
- Swagger documentation for a clear understanding of API endpoints.

## Run standalone (from root directory)

```bash
make dev-dns
```

## API Documentation

API documentation is available via Swagger at the /docs endpoint, e.g., http://localhost:8080/docs/index.html

## End to end example
Check end to end test in [here](./../e2e/dns_test.go)