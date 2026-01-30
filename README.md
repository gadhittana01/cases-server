# Cases App - Backend

Go-based REST API backend for the Legal Marketplace platform. Handles authentication, case management, quote submissions, payment processing via Stripe, and file storage via Supabase Storage (S3-compatible API).

## 🚀 Tech Stack

- **Go 1.24+** - Programming language
- **Gin** - Web framework
- **PostgreSQL** (Supabase) - Database
- **SQLC** - SQL to Go code generation
- **Wire** - Dependency injection
- **golang-migrate** - Database migrations
- **Stripe** - Payment processing
- **Supabase Storage (S3-compatible API)** - File storage via AWS S3 SDK
- **Pusher** - Real-time event broadcasting

## 📋 Prerequisites

- Go 1.24+ installed
- PostgreSQL database (Supabase)
- Supabase Storage bucket with S3-compatible API credentials
- Stripe account (test mode)
- Wire installed: `go install github.com/google/wire/cmd/wire@latest`
- SQLC installed: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

## 🛠️ Setup Instructions

1. **Navigate to server directory:**
   ```bash
   cd server
   ```

2. **Copy environment file:**
   ```bash
   cp config/app.env.example config/app.env
   ```

3. **Update `config/app.env` with your values:**
   ```env
   DB_CONN_STRING=postgresql://user:password@host:port/dbname?sslmode=require
   DB_NAME=your_database_name
   MIGRATION_URL=file://db/migration
   PORT=8000
   JWT_SECRET=your-secret-key
   STRIPE_PUBLISHABLE_KEY=pk_test_...
   STRIPE_SECRET_KEY=sk_test_...
   STRIPE_WEBHOOK_SECRET=whsec_...
   STORAGE_ENDPOINT=https://your-project.supabase.co/storage/v1/s3
   STORAGE_REGION=ap-southeast-1
   STORAGE_ACCESS_KEY=your_access_key
   STORAGE_SECRET_KEY=your_secret_key
   STORAGE_BUCKET=your-bucket-name
   PUSHER_APP_ID=your_app_id
   PUSHER_KEY=your_pusher_key
   PUSHER_SECRET=your_pusher_secret
   PUSHER_CLUSTER=ap1
   FRONTEND_URL=http://localhost:3000
   ```

4. **Install dependencies:**
   ```bash
   go mod tidy
   cd ../go-modules && go mod tidy && cd ../server
   ```

5. **Generate SQLC code:**
   ```bash
   sqlc generate
   ```

6. **Generate Wire dependency injection:**
   ```bash
   go generate ./wire.go
   ```
   Or manually:
   ```bash
   wire ./...
   ```

7. **Run migrations:**
   Migrations run automatically on server start, or manually:
   ```bash
   migrate -path db/migration -database "$DB_CONN_STRING" up
   ```

8. **Start the server:**
   ```bash
   go run main.go wire_gen.go injector.go
   ```

   Server will start on `http://localhost:8000`

## 📁 Project Structure

```
server/
├── api/                      # Vercel serverless functions
│   └── index.go
├── config/                   # Configuration files
│   ├── app.env              # Environment variables
│   └── app.env.example      # Environment template
├── db/
│   ├── migration/           # Database migrations
│   ├── queries/             # SQL queries for sqlc
│   │   ├── cases.sql
│   │   └── quotes.sql
│   └── repository/          # Generated repository code
│       ├── cases.sql.go
│       ├── quotes.sql.go
│       └── repository.go
├── dto/                     # Data transfer objects
│   └── response.go
├── handler/                 # HTTP handlers
│   ├── case_handler.go
│   ├── file_handler.go
│   ├── marketplace_handler.go
│   ├── payment_handler.go
│   ├── quote_handler.go
│   ├── user_handler.go
│   └── webhook_handler.go
├── providers/               # External client providers
│   └── providers.go        # S3, Pusher clients
├── routes/                  # Route definitions
│   └── routes.go
├── service/                 # Business logic
│   ├── case_service.go
│   ├── file_service.go
│   ├── marketplace_service.go
│   ├── payment_service.go
│   ├── quote_service.go
│   └── user_service.go
├── main.go                  # Application entry point
├── wire.go                  # Wire dependency injection config
├── wire_gen.go             # Generated Wire code
└── injector.go             # Dependency injection setup
```

## 📝 API Endpoints

### Public Endpoints

- `POST /api/v1/auth/signup/client` - Client registration
- `POST /api/v1/auth/signup/lawyer` - Lawyer registration
- `POST /api/v1/auth/login` - Login
- `POST /webhooks/stripe` - Stripe webhook handler

### Client Endpoints (Protected, requires `client` role)

- `GET /api/v1/client/cases` - List my cases
- `POST /api/v1/client/cases` - Create case
- `GET /api/v1/client/cases/:id` - Get case details
- `POST /api/v1/client/cases/:id/files` - Upload file
- `POST /api/v1/client/quotes/accept` - Accept quote and create payment intent

### Lawyer Endpoints (Protected, requires `lawyer` role)

- `GET /api/v1/lawyer/marketplace` - List open cases (anonymized)
- `GET /api/v1/lawyer/marketplace/cases/:id` - Get case for marketplace
- `GET /api/v1/lawyer/marketplace/cases/:id/quotes/my` - Get my quote for case
- `POST /api/v1/lawyer/marketplace/cases/:id/quotes` - Submit quote
- `PUT /api/v1/lawyer/marketplace/cases/:id/quotes` - Update quote
- `GET /api/v1/lawyer/quotes` - List my quotes

### Shared Endpoints (Protected)

- `GET /api/v1/auth/profile` - Get current user profile
- `GET /api/v1/files/:id/download` - Get secure download URL

## 🔐 Security Features

1. **Role-Based Access Control (RBAC)**
   - Clients can only access their own cases
   - Lawyers can only see anonymized marketplace cases
   - File access restricted to case owner or accepted lawyer

2. **File Upload Security**
   - Only PDF and PNG files accepted
   - Maximum 10 files per case
   - File size validation (10MB limit)
   - Secure filename generation

3. **Data Anonymization**
   - Client identity hidden in marketplace listings
   - Email/phone redaction in descriptions
   - Full case details only visible after quote acceptance and payment

4. **Payment Security**
   - Atomic quote acceptance (transaction-based)
   - Payment amount validation
   - Stripe PaymentIntent for secure processing
   - Webhook signature verification

5. **Authentication**
   - JWT-based authentication
   - Password hashing with bcrypt
   - Protected routes with role checks

## 🔄 Database Schema

### Key Tables

- **users** - User accounts (clients and lawyers)
- **cases** - Legal cases posted by clients
- **case_files** - Files attached to cases
- **quotes** - Quotes submitted by lawyers
- **payments** - Payment records linked to quotes

### Key Constraints

- One quote per lawyer per case (UNIQUE constraint on case_id + lawyer_id)
- Case status: `open`, `engaged`, `closed`, `cancelled`
- Quote status: `proposed`, `accepted`, `rejected`
- Payment status: `pending`, `succeeded`, `failed`, `canceled`

## 🌐 Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DB_CONN_STRING` | PostgreSQL connection string | Yes |
| `DB_NAME` | Database name | Yes |
| `MIGRATION_URL` | Migration files path | Yes |
| `PORT` | Server port | No (default: 8000) |
| `JWT_SECRET` | Secret for JWT signing | Yes |
| `STRIPE_PUBLISHABLE_KEY` | Stripe publishable key | Yes |
| `STRIPE_SECRET_KEY` | Stripe secret key | Yes |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook secret | Yes |
| `STORAGE_ENDPOINT` | Supabase Storage S3 endpoint | Yes |
| `STORAGE_REGION` | Storage region | Yes |
| `STORAGE_ACCESS_KEY` | Storage access key | Yes |
| `STORAGE_SECRET_KEY` | Storage secret key | Yes |
| `STORAGE_BUCKET` | Storage bucket name | Yes |
| `PUSHER_APP_ID` | Pusher app ID | Yes |
| `PUSHER_KEY` | Pusher app key | Yes |
| `PUSHER_SECRET` | Pusher secret | Yes |
| `PUSHER_CLUSTER` | Pusher cluster | Yes |
| `FRONTEND_URL` | Frontend URL for CORS | Yes |

## 🚢 Deployment

### Vercel Deployment (Serverless)

See detailed guide: [`VERCEL_DEPLOYMENT.md`](VERCEL_DEPLOYMENT.md)

**Quick Steps:**
1. Push your code to GitHub/GitLab/Bitbucket
2. Import project in [Vercel Dashboard](https://vercel.com/dashboard)
3. **Don't set a root directory** (deploy from repository root)
4. Add all environment variables in Vercel
5. Deploy!

**Note**: Vercel uses serverless functions, which have timeout limits. For long-running operations, consider Railway, Render, or Fly.io.

### Traditional Deployment (Docker/VM)

1. Set environment variables in your hosting platform
2. Build the binary:
   ```bash
   go build -o server main.go wire_gen.go injector.go
   ```
3. Run migrations
4. Start the server

### Supabase Storage Setup

1. Create a storage bucket in Supabase (e.g., `my-bucket`)
2. Get S3-compatible API credentials from Supabase Storage settings
3. Update storage-related environment variables in your `.env` file
4. Set bucket to private

## 🧪 Testing

### Running Tests

```bash
go test ./service/... -v
```

### Test Accounts

- **Client:** `client1@example.com` / `Passw0rd!`
- **Lawyer:** `lawyer1@example.com` / `Passw0rd!`

> **Note:** These accounts need to be created through the signup flow.

## 🔧 Development

### Regenerating Code

**SQLC (after changing SQL queries):**
```bash
sqlc generate
```

**Wire (after changing dependencies):**
```bash
go generate ./wire.go
```

### Database Migrations

Migrations are located in `db/migration/` and run automatically on server start. To run manually:

```bash
migrate -path db/migration -database "$DB_CONN_STRING" up
```

To create a new migration:
```bash
migrate create -ext sql -dir db/migration -seq migration_name
```

## 🐛 Troubleshooting

### Database Connection Issues
- Verify connection string format
- Check if database exists
- Ensure network access is allowed
- For `EOF` errors, ensure using `pgx/v5/stdlib` driver (already configured)

### File Upload Issues
- Verify Supabase Storage bucket exists
- Check S3-compatible API credentials (endpoint, access key, secret key)
- Ensure bucket permissions are correct
- Verify the storage endpoint URL format is correct

### Stripe Payment Issues
- Verify Stripe keys are correct
- Check Stripe dashboard for payment intents
- Ensure webhook endpoints are configured correctly
- Verify webhook signature secret matches

### Dependency Injection Issues
- Ensure `wire_gen.go` is regenerated after changes
- Check that all providers are properly wired
- Verify `injector.go` is up to date

## 📄 License

This project is created for the Sibyl Full-Stack Technical Test.
