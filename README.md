# Rinha de Backend 2024/Q1 - NGINX/GO/FIBER/GORM/Postgres

# Stack
- `Nginx 1.25` Load balance
- `PostgreSQL 16.2` Banco de dados
- `Go 1.22` + [GO Fiber](https://gofiber.io/) API

# Rodando aplicação
Dev mode

```
docker-compose -f docker-compose.dev.yml up -d

curl --location 'http://127.0.0.1:9999/clientes/1/extrato'
```

Completo

```
docker-compose up -d --build

curl --location 'http://127.0.0.1:9999/clientes/1/extrato'
```

# Contato
- Linkedin: [Raphael de Monticello](https://www.linkedin.com/in/monticello/)


# Gatling
<img width="1018" alt="Screenshot 2024-02-25 at 3 41 17 AM" src="https://github.com/raphaelmonticello/rinha-backend-2024-q1-gormpg/assets/103146228/dc7062a2-6a87-4d60-bf93-d41cae748cdb">
