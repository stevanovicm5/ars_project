---- 28. Septembar 2025 ----
Opis:
Potpuna CRUD Implementacija i Arhitektonski Refaktor
Ova faza je fokusirana na finalizaciju CRUD funkcionalnosti i optimizaciju strukture koda, čime je projekat spreman za perzistentno skladištenje (Consul).

Ključne Promene:
Finalizovan CRUD: Implementiran je kompletan CRUD (Add, Get, Update, Delete) za Konfiguracione Grupe i usklađeni su svi Handleri i metode repozitorijuma.

Refaktor Rutiranja: Stari switch ruter zamenjen je gorilla/mux-om, a logika je izdvojena u setupRouter() za čistu main funkciju.

Standardizacija: Uklonjena je nekonzistentnost u Handlerima (npr. standardizovani su nazivi Add/Get/Update/Delete).

Unit Testovi: Dodati su robusti Unit Testovi za InMemoryRepository, čime je potvrđena ispravnost kompletne CRUD logike za Konfiguracije i Grupe.

Sledeći korak: Implementacija ConsulRepository.