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