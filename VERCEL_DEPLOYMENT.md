# Vercel Deployment Guide for Go Backend

## ⚠️ Important Considerations

**Vercel is primarily designed for serverless functions**, which means:

1. **Cold Starts**: First request after inactivity may be slower (1-3 seconds)
2. **Timeout Limits**: 
   - Free tier: 10 seconds
   - Pro tier: 60 seconds
   - Enterprise: 300 seconds
3. **Database Connections**: Use connection pooling (already implemented)
4. **File Uploads**: Large file uploads may hit timeout limits
5. **Webhooks**: Stripe webhooks should work, but ensure they complete within timeout

**Alternative Platforms** (better for full Go servers):
- **Railway** - Easy deployment, supports Docker
- **Render** - Free tier available, Docker support
- **Fly.io** - Global edge deployment
- **AWS/GCP/Azure** - Full control, more complex setup

## Prerequisites

1. Vercel account (sign up at https://vercel.com)
2. GitHub/GitLab/Bitbucket repository
3. All environment variables ready
4. Database accessible from Vercel's IPs (Supabase allows this)

## Deployment Steps

### 1. Prepare Your Code

The serverless function handler is already created at `server/api/index.go`. This wraps your Gin router to work with Vercel's serverless functions.

### 2. Handle Monorepo Structure

**Important**: Your project uses a monorepo with `go-modules` as a local dependency. You have two options:

#### Option A: Deploy from Root (Recommended)

1. Keep root directory as repository root (don't set a root directory)
2. Update `vercel.json` to point to `server/api/index.go`
3. Vercel will have access to both `server/` and `go-modules/` directories

#### Option B: Copy go-modules into server

1. Copy `go-modules` into `server/go-modules`
2. Update `go.mod` replace directive to `./go-modules`
3. Set root directory to `server`

**We'll use Option A** - deploy from root with updated paths.

### 3. Connect Repository to Vercel

1. Go to [Vercel Dashboard](https://vercel.com/dashboard)
2. Click "Add New Project"
3. Import your Git repository
4. **Don't set a root directory** (leave it as repository root)

### 4. Configure Project Settings

- **Framework Preset**: Other (or leave blank)
- **Root Directory**: Leave empty (use repository root) ⚠️ **Important**
- **Build Command**: Leave empty (Vercel auto-detects Go)
- **Output Directory**: Leave empty
- **Install Command**: Leave empty

**Note**: The `vercel.json` file is configured to use `server/api/index.go` as the entry point, so we deploy from the repository root to access both `server/` and `go-modules/` directories.

### 5. Set Environment Variables

In Vercel Dashboard → Your Project → Settings → Environment Variables, add:

#### Required Variables:

```
DB_CONN_STRING=postgresql://...
DB_NAME=cases_db
MIGRATION_URL=file://db/migration
JWT_SECRET=your-secret-key
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STORAGE_ENDPOINT=https://...
STORAGE_REGION=ap-southeast-1
STORAGE_ACCESS_KEY=...
STORAGE_SECRET_KEY=...
STORAGE_BUCKET=my-bucket
PUSHER_APP_ID=2108509
PUSHER_KEY=...
PUSHER_SECRET=...
PUSHER_CLUSTER=ap1
```

#### Optional Variables:

```
PORT=8000
```

**Note**: All environment variables are available to your Go code via `os.Getenv()`.

### 5. Handle Database Migrations

**Important**: In serverless environments, migrations should be run separately or on first cold start.

**Option 1: Run migrations before deployment**
```bash
# Run migrations locally or via CI/CD before deploying
migrate -path db/migration -database "$DB_CONN_STRING" up
```

**Option 2: Auto-migration on cold start** (current implementation)
- Migrations run automatically on first request (cold start)
- This may slow down the first request
- Subsequent requests reuse the same instance

**Option 3: Use a migration service**
- Run migrations via a separate service or CI/CD pipeline
- Disable auto-migration in production

### 6. Deploy

1. Click "Deploy"
2. Vercel will:
   - Detect Go files
   - Build the serverless function
   - Deploy to a unique URL

### 7. Verify Deployment

1. Check deployment logs for any errors
2. Test health endpoint: `https://your-project.vercel.app/health`
3. Test API endpoints
4. Monitor function logs in Vercel dashboard

## Project Structure for Vercel

```
cases-app/                 # Repository root (Vercel root directory)
├── vercel.json            # Vercel configuration (at root)
├── server/
│   ├── api/
│   │   └── index.go       # Vercel serverless function handler
│   ├── db/
│   │   └── migration/     # Migration files (included in deployment)
│   ├── handler/
│   ├── service/
│   ├── routes/
│   └── ... (other files)
└── go-modules/            # Shared Go utilities (accessible from root)
    ├── middleware/
    └── utils/
```

## Environment Variables Reference

| Variable | Description | Required |
|----------|-------------|----------|
| `DB_CONN_STRING` | PostgreSQL connection string | ✅ |
| `DB_NAME` | Database name | ✅ |
| `MIGRATION_URL` | Migration files path | ✅ |
| `JWT_SECRET` | Secret for JWT signing | ✅ |
| `STRIPE_PUBLISHABLE_KEY` | Stripe publishable key | ✅ |
| `STRIPE_SECRET_KEY` | Stripe secret key | ✅ |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook secret | ✅ |
| `STORAGE_ENDPOINT` | Supabase Storage endpoint | ✅ |
| `STORAGE_REGION` | Storage region | ✅ |
| `STORAGE_ACCESS_KEY` | Storage access key | ✅ |
| `STORAGE_SECRET_KEY` | Storage secret key | ✅ |
| `STORAGE_BUCKET` | Storage bucket name | ✅ |
| `PUSHER_APP_ID` | Pusher app ID | ✅ |
| `PUSHER_KEY` | Pusher key | ✅ |
| `PUSHER_SECRET` | Pusher secret | ✅ |
| `PUSHER_CLUSTER` | Pusher cluster | ✅ |
| `PORT` | Server port (not used in serverless) | ❌ |

## Important Notes

1. **Serverless Functions**: Each request spawns a function instance (reused for warm requests)
2. **Cold Starts**: First request after inactivity may take 1-3 seconds
3. **Database Connections**: Connection pooling is handled automatically
4. **File Uploads**: Large files may hit timeout limits (consider using direct upload to Supabase)
5. **Webhooks**: Ensure Stripe webhooks complete within timeout limits
6. **Environment Variables**: Changes require redeployment

## Troubleshooting

### Build Fails?

1. **Check Go version**: Ensure `vercel.json` specifies correct Go version
2. **Check imports**: Ensure all dependencies are in `go.mod`
3. **Check build logs**: Look for specific errors in Vercel dashboard

### Function Timeout?

1. **Upgrade to Pro**: Pro tier has 60-second timeout
2. **Optimize queries**: Ensure database queries are fast
3. **Use background jobs**: For long-running tasks, use queues

### Database Connection Issues?

1. **Check connection string**: Verify format is correct
2. **Check network access**: Ensure database allows Vercel IPs
3. **Check connection pooling**: Ensure pooling is configured correctly

### Migrations Not Running?

1. **Run manually**: Execute migrations before deployment
2. **Check logs**: Look for migration errors in function logs
3. **Disable auto-migration**: If causing issues, disable in code

## Alternative: Deploy as Docker Container

If serverless functions don't meet your needs, consider deploying as a Docker container:

### Railway (Recommended)

1. Push code to GitHub
2. Connect to Railway
3. Railway auto-detects Dockerfile
4. Set environment variables
5. Deploy!

### Render

1. Create new Web Service
2. Connect GitHub repository
3. Set build command: `docker build -t app .`
4. Set start command: `docker run app`
5. Set environment variables
6. Deploy!

## Continuous Deployment

Vercel automatically deploys:
- **Production**: On push to main/master branch
- **Preview**: On every push to other branches/PRs

Each preview gets its own URL for testing.

## Custom Domain

1. Go to Project Settings → Domains
2. Add your custom domain
3. Configure DNS records as instructed
4. SSL is automatically provisioned by Vercel

## Monitoring

- **Function Logs**: Available in Vercel dashboard
- **Metrics**: Request count, duration, errors
- **Alerts**: Set up alerts for errors or high latency
