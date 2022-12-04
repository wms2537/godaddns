# GoDaddy DDNS Service
How To Use
```
docker build -t registry.jomluz.com/ddns:latest .
docker push registry.jomluz.com/ddns:latest
```
Inspired By [https://github.com/proprietary/godaddns](https://github.com/proprietary/godaddns)

## Docker Compose Use
```yaml
version: '3.9'

services:
  ddns:
    image: registry.jomluz.com/ddns
    environment:
      - API_KEY=/run/secrets/domain_api_key
      - API_SECRET=/run/secrets/domain_api_secret
      - SUBDOMAIN=jetson
    secrets:
     - domain_api_key
     - domain_api_secret
secrets:
  domain_api_key:
    file: key.txt
  domain_api_secrets:
    file: secret.txt
```