# Health Bot

Bot personal de WhatsApp para tracking de salud y fitness. Mandás mensajes en lenguaje natural y el bot los interpreta con un LLM, guarda los datos en PostgreSQL y responde con un resumen. Un dashboard web visualiza el progreso.

---

## Estado del proyecto

El core funcional del spec está implementado end-to-end (bot → LLM → DB → API → dashboard). Falta principalmente el deploy real y features de conveniencia.

**Implementado**
- [x] Bot de WhatsApp (whatsmeow) con login por QR, sesión persistida y filtro por grupo + `IsFromMe`
- [x] Parser Groq con fallback a modelo grande cuando `confidence == "low"`
- [x] Schema de DB y migraciones (6 tablas, según spec)
- [x] Queries sqlc para food, strength, cardio, water y body_weight
- [x] API REST con los 5 endpoints (`/today`, `/progress`, `/gym`, `/gym/:exercise`, `/body-weight`)
- [x] Dashboard Next.js con las 3 vistas (Hoy, Progreso, Gym) conectadas a la API, con gráficos (Recharts)
- [x] `docker-compose.yml` con los 5 servicios (postgres, bot, api, web, nginx) + nginx con Basic Auth
- [x] Herramienta `groqtest` para probar el parser sin pasar por WhatsApp

**Pendiente**
- [ ] Deploy real en VPS (todo lo anterior corre local/Docker, no probado en producción)
- [ ] Notificaciones diarias (resumen de calorías al final del día)
- [ ] Comando `/resumen` para ver el día sin abrir el dashboard
- [ ] Editar/eliminar registros ("borra la última comida")
- [ ] Múltiples usuarios
- [ ] Export de datos a CSV

---

## Stack

| Componente | Tecnología |
|---|---|
| Backend | Go 1.26 |
| WhatsApp | [whatsmeow](https://github.com/tulir/whatsmeow) |
| LLM | Groq API (`groq/compound-mini` / `groq/compound`) |
| Base de datos | PostgreSQL 16 |
| Driver DB | pgx/v5 |
| SQL codegen | sqlc |
| Migraciones | golang-migrate |
| HTTP API | Gin |
| Frontend | Next.js 15 (App Router) + Tailwind + Recharts |
| Auth web | HTTP Basic Auth vía nginx |
| Deployment | Docker Compose |
| Timezone | America/Argentina/Buenos_Aires |

---

## Estructura del proyecto

```
health-bot/
├── cmd/
│   ├── bot/main.go          # Entry point del bot de WhatsApp
│   ├── api/main.go          # Entry point de la API REST
│   └── groqtest/main.go     # Herramienta de prueba del parser LLM
├── internal/
│   ├── bot/
│   │   ├── handler.go       # Router de mensajes → DB
│   │   └── response.go      # Helpers de formateo de respuestas
│   ├── groq/
│   │   └── parser.go        # Cliente Groq → struct tipado
│   ├── db/
│   │   ├── connect.go       # Conexión pgxpool
│   │   ├── db.go            # Generado por sqlc
│   │   ├── models.go        # Generado por sqlc
│   │   └── *.sql.go         # Queries generadas por sqlc
│   └── api/
│       ├── routes.go        # Registro de rutas Gin
│       └── handlers.go      # Endpoints del dashboard
├── web/                     # Next.js App Router
│   └── app/
│       ├── page.tsx         # Vista: Hoy
│       ├── progress/        # Vista: Progreso
│       └── gym/             # Vista: Gym
├── migrations/              # SQL up/down para golang-migrate
├── queries/                 # SQL fuente para sqlc
├── sqlc.yaml
├── docker-compose.yml
├── nginx.conf
├── Dockerfile.bot
└── Dockerfile.api
```

---

## Cómo funciona

### Bot de WhatsApp

El bot corre con el número personal del usuario vía whatsmeow (linked device). Escucha **solo** los mensajes enviados por el propio usuario en un grupo dedicado.

```
Mensaje entrante (IsFromMe=true, grupo configurado)
        ↓
groq/parser.go → LLM → JSON estructurado
        ↓
Según tipo: guardar en tabla correspondiente
        ↓
Responder con resumen al chat
```

### Tipos de mensaje soportados

| Ejemplo | Tipo | Respuesta |
|---|---|---|
| "almorcé milanesa con puré" | food | `✅ Almuerzo · +680 kcal (P:45g C:60g G:22g F:4g) · Quedan 1.220 kcal` |
| "hice press de banca 4x8 a 80kg" | strength | `✅ Bench press 4×8 @80.0kg · -150 kcal quemadas` |
| "corrí 5km en 28 minutos" | cardio | `✅ Running 5.0km / 28min · -320 kcal quemadas` |
| "peso 78.5kg" | body_weight | `✅ Peso registrado: 78.5kg` |
| "tomé 1.5 litros de agua" | water | `✅ +1.5L · Total hoy: 2.5L` |

### Parser LLM (Groq)

- Modelo rápido: `groq/compound-mini`
- Modelo fallback: `groq/compound` (se usa si `confidence == "low"`)
- Timeout HTTP: 30 segundos
- Normaliza nombres de ejercicios a inglés ("press de banca" → "bench press")
- Estima macros y calorías cuando no se especifican

---

## Schema de base de datos

```sql
users           -- configuración del usuario (calorie_goal, protein_goal_g)
food_entries    -- comidas con macros (calories, protein_g, carbs_g, fat_g, fiber_g)
strength_sessions -- ejercicios de fuerza (exercise_name, sets, reps, weight_kg)
cardio_sessions -- cardio (activity, duration_min, distance_km, calories_burned)
body_weight     -- historial de peso corporal
water_logs      -- hidratación diaria
```

---

## API REST

Base URL: `/api`

| Endpoint | Descripción |
|---|---|
| `GET /api/today` | Resumen del día: calorías, macros, agua, entrenamientos |
| `GET /api/progress?month=YYYY-MM` | Tendencias mensuales + adherencia |
| `GET /api/gym` | Lista de ejercicios registrados |
| `GET /api/gym/:exercise` | Evolución de peso + PRs por mes |
| `GET /api/body-weight` | Historial de peso corporal |

---

## Dashboard web

### Vista: Hoy (`/`)
- Barra de progreso calórico (consumidas vs objetivo)
- Breakdown de macros: Proteínas / Carbos / Grasas / Fibra
- Agua consumida del día
- Lista de comidas con calorías por ítem
- Lista de entrenamientos del día

### Vista: Progreso (`/progress`)
- Gráfico de línea: peso corporal en el tiempo
- Gráfico de barras: calorías diarias vs objetivo
- Métrica: adherencia mensual (% de días en objetivo)

### Vista: Gym (`/gym`)
- Selector de ejercicio
- Gráfico de línea: evolución del peso levantado
- Tabla de PRs por mes

---

## Setup local

### Requisitos
- Go 1.26+
- Docker Desktop
- Node.js 22+
- [sqlc](https://sqlc.dev) (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
- [golang-migrate](https://github.com/golang-migrate/migrate) (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)

### 1. Variables de entorno

```bash
cp .env.example .env
# Editar .env con tus valores
```

```env
GROQ_API_KEY=gsk_...
WHATSAPP_GROUP_JID=120363XXXXXXXXXX@g.us
WHATSAPP_SESSION_DIR=./session
DATABASE_URL=postgres://user:password@localhost:5432/healthbot?sslmode=disable
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=healthbot
API_PORT=8080
TZ=America/Argentina/Buenos_Aires
```

### 2. Base de datos

```bash
docker compose up postgres -d

migrate -path migrations \
  -database "postgres://user:password@localhost:5432/healthbot?sslmode=disable" up
```

### 3. Primer login de WhatsApp

La primera vez hay que escanear un QR para vincular el dispositivo:

```bash
export GROQ_API_KEY=...
export DATABASE_URL=...
export WHATSAPP_SESSION_DIR=./session
export WHATSAPP_GROUP_JID=placeholder   # cualquier valor por ahora
export USER_PHONE=tuNumero
go run ./cmd/bot/
```

Escaneá el QR desde **WhatsApp → Dispositivos vinculados → Vincular un dispositivo**.

La sesión queda guardada en `./session/` — los reinicios posteriores no necesitan QR.

### 4. Obtener el JID del grupo

Creá un grupo dedicado en WhatsApp. Una vez conectado el bot, mandá cualquier mensaje al grupo y el JID aparece en los logs:

```
[JID] Message in group: 120363411941801113@g.us
```

Actualizá `WHATSAPP_GROUP_JID` en el `.env` con ese valor.

### 5. Correr el bot

```bash
export $(cat .env | grep -v '#' | xargs)
go run ./cmd/bot/
```

### 6. Correr la API

```bash
export $(cat .env | grep -v '#' | xargs)
go run ./cmd/api/
```

### 7. Correr el dashboard

```bash
cd web
npm install
npm run dev
```

---

## Deployment (Docker Compose)

> La sesión de WhatsApp debe estar inicializada localmente antes del primer deploy.

```bash
docker compose up -d
```

Servicios:
- **postgres** — PostgreSQL 16 con volumen persistente
- **bot** — bot de WhatsApp, monta `./session/`
- **api** — API REST en `:8080`
- **web** — Next.js en `:3000`
- **nginx** — reverse proxy con HTTP Basic Auth en `:80`

Para generar el `.htpasswd` de nginx:
```bash
htpasswd -c .htpasswd tuUsuario
```

---

## Probar el parser de Groq

```bash
export GROQ_API_KEY=gsk_...
go run ./cmd/groqtest/
```

Envía 5 mensajes de ejemplo (con delay entre ellos por rate limit del free tier) y muestra el JSON parseado.

---

## Decisiones técnicas

- **Sin comandos en WhatsApp** — todo lenguaje natural, el LLM clasifica
- **whatsmeow** — linked device en Go, sin Puppeteer, bajo consumo de RAM
- **sqlc sobre ORM** — SQL puro generado a código Go type-safe
- **modernc.org/sqlite** — driver SQLite puro en Go (sin CGO) para la sesión de whatsmeow
- **WAL mode en SQLite** — evita SQLITE_BUSY en escrituras concurrentes de whatsmeow
- **Fechas como `DATE`** — sin timezone issues, día calculado en UTC-3 en la app
- **Ejercicios en inglés** — el LLM normaliza para queries consistentes en el dashboard
- **Un solo usuario** — no hay auth en la API, nginx protege la web con Basic Auth
- **Groq fallback** — si `confidence == "low"`, reintenta con modelo más grande

Ver [Estado del proyecto](#estado-del-proyecto) arriba para el detalle de qué falta.
