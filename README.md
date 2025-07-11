# Cashout AI

Telegram **AI Agent** for Income and Expense Management with **Web Dashboard**.

You can self-host it following the Developer section down below.

<p align="center">
  <img src="/assets/demo.gif" alt="Demo" height="550px">
</p>

## Features

Cashout is an intelligent Telegram bot that leverages AI to make expense tracking effortless. Simply send a message in natural language, and the bot will understand and categorize your transactions automatically.

### 🤖 AI-Powered Transaction Processing

- **Natural Language Understanding**: Just type "coffee 3.50" or "salary 3000 yesterday" - no complex commands needed
- **Smart Categorization**: Automatically assigns the right category based on your description
- **Flexible Date Recognition**: Understands various date formats (dd/mm, dd-mm-yyyy, "yesterday", etc.)
- **Multi-language Support**: Works with transaction descriptions in any language

### 💰 Transaction Management

- **Quick Entry**: Add expenses and income with a single message
- **Inline Editing**: Modify amount, category, description, or date before confirming
- **Bulk Operations**: Edit or delete existing transactions with paginated navigation
- **Transaction Types**: Track both expenses (18 categories) and income (2 categories)
- **Search and Full Listing**: Find transactions by full text search and category or full listing
- **Export Functionality**: Download all your transactions as CSV files

### 📊 Financial Insights

- **Weekly Recap**: Get detailed breakdowns of your current week's spending
- **Monthly Summary**: View month-by-month financial performance with category breakdowns
- **Yearly Overview**: See annual trends and top spending categories
- **Balance Tracking**: Instant calculation of income vs expenses for any period
- **Category Analysis**: Understand where your money goes with percentage breakdowns

### 🌐 Web Dashboard

- **Secure Authentication**: Telegram-based login with verification codes
- **Monthly Views**: Navigate through different months with intuitive controls
- **Visual Insights**: Clear categorization and trend analysis

### 🔔 Smart Reminders

- **Automated Weekly Recaps**: Receive your previous week's summary every Monday
- **Automated Monthly Recaps**: Receive your previous month's summary on the 1st of each month
- **Intelligent Scheduling**: Only sends reminders to active users
- **Reliable Delivery**: Built-in retry mechanism for failed notifications

### 💻 Available Commands

- `/start` - Initialize the bot and see the main menu
- `/edit` - Edit an existing transaction
- `/delete` - Delete a transaction
- `/list` - View all transactions (paginated)
- `/search` - Search transactions by description
- `/week` - Get current week's financial summary
- `/month` - Get current month's financial summary
- `/year` - Get current year's financial summary
- `/export` - Export all transactions to CSV

### 🎯 User Experience

- **Intuitive Interface**: Clean inline keyboards for all operations
- **Smart Navigation**: Year/month selectors for browsing historical data
- **Pagination**: Handle large transaction lists with ease
- **Quick Actions**: Home screen with instant access to all major functions
- **Cancel Anytime**: Every operation can be cancelled mid-flow

### 🛠️ Technical Features

- **Database Migrations**: Version-controlled schema management
- **Webhook & Polling Support**: Flexible deployment options
- **Development Tools**: Built-in database seeder for testing
- **Modular Architecture**: Clean separation of concerns for easy maintenance
- **Configurable Access**: Optional user whitelist for private deployments

## Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL Database
- Access to an OpenAI-compatible API model, with its API Key and Endpoint (e.g. DeepSeek, OpenAI, etc.)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/alainrk/cashout.git
cd cashout
```

2. Install dependencies:

```bash
go mod download
```

### Environment Setup

```bash
cp .env.example .env
```

Copy the example `.env` file in the project root (or set environment variables) and edit it accordingly:

```env
TELEGRAM_BOT_API_TOKEN='XXXXXXXXXX:AAAA_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbb'
DATABASE_URL='postgres://postgres:postgres@localhost:5432/postgres'
OPENAI_API_KEY='sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'
OPENAI_BASE_URL='https://api.deepseek.com/v1'
LLM_MODEL='deepseek-chat'
RUN_MODE='webhook' # webhook or polling
WEBHOOK_DOMAIN=''
WEBHOOK_SECRET=''
WEBHOOK_HOST='localhost'
WEBHOOK_PORT='8080'
LOG_LEVEL='info'
# Dev purpose, comma separated. Keep it empty to allow all
ALLOWED_USERS=''
# Seed purpose - set the Telegram ID of the user to seed transactions for
SEED_USER_TG_ID=''
# Web Server Configuration
WEB_HOST=localhost
WEB_PORT=8081
# Session Configuration (optional)
SESSION_SECRET=your-random-session-secret-here
SESSION_DURATION=24h
```

Spin up local infrastructure:

```bash
docker compose up -d
```

## Database Management

Cashout uses a version-based migration system to manage database schema changes.

Use the following commands to manage database migrations:

```bash
# Run all pending migrations
go run ./cmd/migrate/main.go -command up

# Run all pending migrations with another .env file
go run ./cmd/migrate/main.go -command up -env .prod.env

# Create a new migration just by copy-pasting a previous one and editing it accordingly
cp internal/migrations/versions/001*.go internal/migrations/versions/00X_your_migration.go
```

For migrations that support rollback, use the `RegisterMigrationWithRollback` function:

```go
func init() {
    migrations.RegisterMigrationWithRollback("003", "Add email column",
        add_email_column, rollback_email_column)
}

func rollback_email_column(tx *gorm.DB) error {
    return tx.Exec(`ALTER TABLE users DROP COLUMN email`).Error
}
```

## Development

### Building and Running

#### Telegram Bot

```bash
# Build the bot
make build

# Run the bot
make run

# Run the bot with live reloading (requires Air)
make run/live
```

#### Web Server

```bash
# Build the web server
make build-web

# Run the web server
make run-web

# Run the web server with live reloading
make run/live-web
```

#### Both Services

```bash
# Build both applications
make build-all

# Build both for Linux
make build-linux-all

# Note: Running both requires two terminals
# Terminal 1: make run
# Terminal 2: make run-web
```

### Database Seeding

The Dev DB Seeder generates test transaction data for development:

```bash
# Set the user's Telegram ID you want to seed data for
export SEED_USER_TG_ID=123456789

# Seed the database with random transactions
make db/seed
```

The seeder will:

- Generate 5 years of transaction history
- Create 90% expenses and 10% income transactions
- Distribute transactions across all categories
- Ensure at least one salary per month
- Delete existing transactions before seeding (idempotent)

## Deployment

### Docker Compose (Recommended)

The project includes a complete Docker Compose setup:

```bash
# Start all services (database, bot, web server)
docker compose up -d

# View logs
docker compose logs -f

# Stop all services
docker compose down
```

This will start:

- PostgreSQL database on port 5432
- Telegram bot (webhook mode on port 8080)
- Web dashboard on port 8081
- Automatic database migrations

### Manual Deployment

#### Telegram Bot

The bot can run both in `webhook` and `polling` mode.

**Webhook Mode:**

```env
RUN_MODE='webhook'
WEBHOOK_DOMAIN='https://your-domain.com'
WEBHOOK_SECRET='xxxyyyzzz'
WEBHOOK_PORT='8080'
```

**Polling Mode:**

```env
RUN_MODE='polling'
```

#### Web Server

The web server runs independently and can be configured:

```env
WEB_HOST=0.0.0.0  # For production
WEB_PORT=8081
SESSION_SECRET=your-random-session-secret-here
SESSION_DURATION=24h
```

### LLM Setup

Any OpenAI compatible API LLM can be used:

**Example with DeepSeek:**

```env
OPENAI_API_KEY='sk-xxx'
OPENAI_BASE_URL='https://api.deepseek.com/v1'
LLM_MODEL='deepseek-chat'
```

**Example with OpenAI:**

```env
OPENAI_API_KEY='sk-xxx'
OPENAI_BASE_URL='https://api.openai.com/v1'
LLM_MODEL='gpt-4'
```

## Web Dashboard Usage

1. **Access**: Navigate to `http://localhost:8081` (or your configured domain)
2. **Login**: Enter your Telegram username
3. **Verification**: Check Telegram for a 6-digit verification code
4. **Dashboard**: View your financial data with month navigation
5. **Statistics**: See real-time balance, income, expenses, and transaction counts
6. **History**: Browse detailed transaction history with search and filtering

The web dashboard provides a complementary interface to the Telegram bot, offering:

- Better visualization for large datasets
- Month-by-month navigation
- Desktop-friendly transaction management
- Exportable financial reports

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests in CI mode
make test-ci

# Run security checks
make sec
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
