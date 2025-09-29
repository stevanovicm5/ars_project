# ars_project

- Milan Stevanović
* Ognjen Karišik
+ Ognjen Burmazović

# Configuration Management API

Ova aplikacija predstavlja REST API za upravljanje konfiguracionim parametrima i grupama konfiguracija. Omogućava centralizovano upravljanje konfiguracijama sa podrškom za idempotentne operacije i versioning.

## Karakteristike

- **CRUD Operacije** - Kompletno upravljanje konfiguracijama i konfiguracionim grupama
- **Idempotentnost** - Garantovano jednokratno izvršavanje operacija putem X-Request-Id headera
- **Dockerizacija** - Potpuno kontejnerizovana aplikacija sa Docker Compose
- **API Dokumentacija** - Automatski generisana Swagger/OpenAPI dokumentacija
- **Consul Integracija** - Persistencija podataka kroz HashiCorp Consul KV store
- **Health Checks** - Integrisani health check endpointi za monitoring

## Korišćene tehnologije

- Golang 1.25.1
- Consul 1.15.4
- Docker

### Pokretanje celokupnog sistema

```
# Kloniranje repozitorijuma
git clone <repository-url>
cd alati_projekat

# Pokretanje svih servisa
docker-compose up --build
```

## Pristup servisima

[Consul](http://localhost:8500)

[Swagger](http://localhost:8080/swagger/index.html)