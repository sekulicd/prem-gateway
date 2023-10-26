# prem-gateway

## Project Description

`prem-gateway` acts as an API gateway for directing and managing a multitude of operations. <br /> 
It is responsible for routing requests from the frontend `prem-app` to either the `prem-daemon` for Docker image management or directly to Docker images providing `prem-services`.

## Features

- [x] API Gateway
- [x] Authentication/Authorization
- [x] Domain Management
- [x] TLS
- [x] Rate Limiting
- [ ] Logging
- [ ] Metrics

## Services

- [dnsd](./dns/README.md)
- [authd](./auth/README.md)
- [controllerd](./controller/README.md)

## Usage
Create network:
```bash
docker network create prem-gateway
```

Change permission:
```bash
chmod 600 ./traefik/letsencrypt/acme.json
```

Start prem-gateway:
```bash
make up 
```
#### Default Let's Encrypt CA server is the staging. For production, start prem-gateway with bellow command'.
```bash
make up LETSENCRYPT_PROD=true
```

Stop prem-gateway:
```bash
make down
```

#### In order to restart services outside prem-gateway and to assign them with subdomain/tls certificate, use bellow command.
```bash
make up LETSENCRYPT_PROD=true
```

#### Run prem-gateway with prem-app and prem-daemon:
```bash
make runall PREMD_IMAGE={IMG} PREMAPP_IMAGE={IMG}
```

#### Stop prem-gateway, prem-app and prem-daemon:
```bash
make stopall PREMD_IMAGE={IMG} PREMAPP_IMAGE={IMG}
```
