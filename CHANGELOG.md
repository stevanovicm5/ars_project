---- 28. Septembar 2025 ----
Ognjen K
Opis:
Potpuna CRUD Implementacija i Arhitektonski Refaktor
Ova faza je fokusirana na finalizaciju CRUD funkcionalnosti i optimizaciju strukture koda, čime je projekat spreman za perzistentno skladištenje (Consul).

Ključne Promene:
Finalizovan CRUD: Implementiran je kompletan CRUD (Add, Get, Update, Delete) za Konfiguracione Grupe i usklađeni su svi Handleri i metode repozitorijuma.

Refaktor Rutiranja: Stari switch ruter zamenjen je gorilla/mux-om, a logika je izdvojena u setupRouter() za čistu main funkciju.

Standardizacija: Uklonjena je nekonzistentnost u Handlerima (npr. standardizovani su nazivi Add/Get/Update/Delete).

Unit Testovi: Dodati su robusti Unit Testovi za InMemoryRepository, čime je potvrđena ispravnost kompletne CRUD logike za Konfiguracije i Grupe.

Sledeći korak: Implementacija ConsulRepository.

Opis:
Implementacija ConsulRepository: Uspešno implementirana logika za CRUD operacije koristeći Consul KV Store.

Stabilizacija Okruženja: Rešeni problemi sa konekcijom Consula kroz Docker i fiksiran nevalidni karakter 

Testna Potvrda (Ocena 8): Svi InMemory i Consul integracioni testovi sada su PASS, potvrđujući punu funkcionalnost repozitorijuma i ispunjavanje uslova za perzistenciju podataka u Consulu.

Opis:
Ova izmena kompletira implementaciju Idempotentnosti (Ocena 9) obezbeđujući semantiku 'Exactly-Once Execution' za sve operacije koje menjaju stanje:
/configurations (POST, PUT)
/configgroups (POST, PUT)

Ključne promene:
1. Ažuriran IdempotencyMiddleware: Uklonjena je logika za 'SaveIdempotencyKey()' koja se izvršavala neselektivno NAKON što se Handler izvrši. Middleware je sada isključivo zadužen za 'Check' i 'Block'.
2. Migracija logike snimanja: Logika 'app.Repo.SaveIdempotencyKey(key)' je premeštena u sva 4 Handler-a (Add/Update za Konfiguracije/Grupe). Ključ se sada čuva u Consulu ISKLJUČIVO ako je repozitorijumska operacija bila uspešna (bez 400 ili 409 grešaka).
3. Ispravljen ConfigHandler: Ažurirana zavisnost da koristi ispravan interfejs 'repository.Repository'.

---- 29. Septembar 2025 ----
Ognjen B
Opis:
Kompletna Dockerizacija i API Dokumentacija
Ova faza je fokusirana na pripremu aplikacije za produkciju kroz kompletnu Dockerizaciju i Swagger dokumentaciju, čime je projekat spreman za inicijalan deployment.

Ključne Implementacije:
1. Swagger API Dokumentacija:
Implementiran kompletan Swagger UI sa detaljnom dokumentacijom svih API endpointa
Generisane automatske OpenAPI specifikacije za sve CRUD operacije
Dokumentovani svi handleri sa odgovarajućim annotation komentarima
Omogućen pristup Swagger interfejsu na /swagger/index.html

2. Docker Konfiguracija:
Kreiran optimizovan Dockerfile sa multi-stage build procesom
Implementiran .dockerignore za efikasnije buildovanje
Konfigurisan docker-compose.yaml za pokretanje kompletnog sistema
Omogućena komunikacija između aplikacije i Consul servisa

3. Strukturna Organizacija:
Premešten main.go u root direktorijum projekta za bolju strukturu
Ažuriran go.mod sa svim potrebnim dependency-ima
Kreiran .gitignore sa optimalnim podešavanjima za Go i Docker projekte

4. Kontejnerizacija Servisa:
Docker Compose konfigurisan sa dva servisa: app i consul
Omogućena mrežna komunikacija između kontejnera
Konfigurisani portovi: 8080 za aplikaciju, 8500 za Consul UI

5. Ažuriranje Dokumentacije:
Kompletan README.md sa uputstvima za pokretanje
Detaljne instrukcije za build i deployment
MIT licenca

---- 29. Septembar 2025 ----
Ognjen K
Opis:
Uvođenje Servisnog Sloja i Metrika (Ocena 10)
Ova faza kompletira prelazak na troslojnu arhitekturu (Handler -> Service -> Repository) i implementira praćenje performansi.

Ključne Implementacije:
1. Uvođenje Servisnog Sloja (`services/`):
- Izdvojena sva poslovna logika (uključujući logiku Idempotentnosti) iz Handlera u novi `ConfigurationService`.
- Svi Handleri sada komuniciraju isključivo sa Servisnim Slojem.

2. Arhitektura Interfejsa:
- Definisan `services.Service` interfejs, čime je omogućena potpuna izolacija slojeva i korišćenje Decorator Pattern-a.

3. Implementacija Metrika (Prometheus/Go-Kit):
- Kreiran `MetricsService` kao **Decorator** oko osnovnog `ConfigurationService`.
- `MetricsService` automatski meri latenciju (`Histogram`) i broji pozive (`Counter`) za sve glavne CRUD operacije (Add, Get, Update, Delete), bez menjanja poslovne logike.

4. Izlaganje Metrika:
- Dodat `/metrics` endpoint koristeći `promhttp.Handler()` za izlaganje Prometheus metrika.

5. Ažuriranje Handlera:
- `handlers.NewConfigHandler` sada prima `services.Service` **interfejs**, čime je rešena zavisnost od konkretne implementacije.

Uspešno smo integrisali i otklonili sve probleme kako bi Vaš sistem pouzdano prikupljao i vizualizovao performanse.

Potvrđena Infrastruktura: Dokazali smo da Docker, Prometheus i Grafana rade i vide se međusobno. Prometheus sada pouzdano čita podatke sa http://app:8080/metrics.

Aktivirane Metrike: Otklonili smo greške u Go kodu (services/metrics_service.go), čime smo osigurali da se brojač app_http_requests_total zaista povećava sa svakim pozivom (npr., AddConfigurationGroup).

Rešen Konflikt Rutiranja: Eliminacijom duplih registracija i izolacijom rute /metrics u main.go, uklonili smo konflikt sa Idempotency Middleware-om. Ovo je bila ključna promena koja je omogućila prometheus hendleru da radi bez smetnji.

Vizualizacija u Grafani: Povezali smo Grafanu sa Prometheusom, što Vam omogućava da prate (vizualizujete) u realnom vremenu stopu poziva (rate(app_http_requests_total[1m])) i druge metrike.

---- 29. Septembar 2025 ----
Milan S
Opis:
Proširenje Funkcionalnosti i Poboljšanje API-ja
Ova faza je fokusirana na dodavanje naprednih funkcionalnosti i poboljšanje developer experience-a kroz kompletno dokumentovanje.

Ključne Implementacije:
1. Proširenje Modela sa Labelima:
Dodati labeli u Configuration model
Labeli kao metadata za kategorizaciju i filtriranje konfiguracija
Key-value struktura za fleksibilno označavanje

2. Napredne Query Operacije:
GET po labelima - Filtriranje konfiguracija prema label kriterijumima
DELETE po labelima - Brisanje konfiguracija na osnovu labela
Fleksibilni query parametri za kompleksno filtriranje

3. Helper Metode za Label Operacije:
Label-based pretrage u repository sloju
Validacija label strukture
Efikasno filtriranje u Consul KV store-u

4. Kompletna Swagger Dokumentacija:
Detaljni OpenAPI opisi za sve endpointove
Primeri request/response za sve operacije
Model dokumentacija sa opisima svih polja
Error response dokumentacija sa HTTP status kodovima

---- 29. Septembar 2025 ----
Ognjen B
Opis:
Refaktor Middleware Arhitekture i Implementacija Rate Limitinga
Ova faza je fokusirana na poboljšanje arhitekture i dodavanje zaštitnih mehanizama za API.

Ključne Promene:
1. Restruktuiranje Middleware Sloja:
Izdvojen IdempotencyMiddleware iz main.go u zaseban fajl middleware/idempotency.go
Kreiran modularan middleware paket sa jasno definisanim responsibilitijem
Poboljšana testabilnost - svaki middleware može se testirati nezavisno

2. Implementacija Rate Limitinga:
Dodat RateLimiter middleware u middleware/ratelimit.go
Konfigurisano ograničenje: 100 zahteva po minuti po IP adresi
Standardni HTTP headeri: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
HTTP 429 status za prekoračenje limita

3. Poboljšana Bezbednost i Stabilnost:
Dodati nil checkovi za sprečavanje panic grešaka
Detaljnije logovanje za debugging i monitoring
Graceful error handling u svim middleware komponentama

4. Ažurirana Docker Konfiguracija:
Middleware fajlovi uključeni u Docker build
Automatsko testiranje rate limitinga pri pokretanju
Poboljšani health checkovi sa statusom limita

---- 30. Septembar 2025 ----
Ognjen K
Opis:
Implementirana distribuirana praćenja (Tracing) sa Jaegerom i OpenTelemetryjem
Kljucne promene:
Implementirano je kompletno praćenje toka zahteva (tracing) za Configuration Service korišćenjem OpenTelemetry SDK-a i OTLP gRPC protokola. Postavili smo Jaeger u Docker Compose za vizualizaciju, što omogućava praćenje latencije i toka podataka kroz sve slojeve aplikacije (Handler, Service, Repository).

---- 1. Oktobar 2025 ----
Ognjen B
Opis:
Dodata Test Infrastruktura i CI/CD Pipeline
Ova faza je fokusirana na implementaciju test strategije i GitHub Actions CI/CD pipeline-a.

Ključne Implementacije:
1. Kompletan Test Suite:
Middleware Testovi - Rate limiting i idempotency middleware
Handler Testovi - HTTP endpointovi sa mock servisima
Service Testovi - Business logika sa mock repository
Repository Testovi - Integracija sa Consul-om

2. GitHub Actions CI Pipeline:
Automatsko pokretanje na push/pull request
Consul kontejner za integracione testove
Multi-stage testiranje - jedinčni i integracioni testovi
Code quality checks - gofmt i go vet