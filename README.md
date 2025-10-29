# Iris - Alert Notification System

Iris is a comprehensive alert management and notification system designed to receive alerts from monitoring systems like Grafana and AlertManager, and dispatch notifications through multiple channels including SMS and email.

## Features

- **Alert Management**: Receive, store, and manage alerts from Grafana and AlertManager
- **Multi-Channel Notifications**: Support for SMS (Kavenegar, Smsir) and Email notifications
- **User & Role Management**: Complete RBAC (Role-Based Access Control) system
- **Group Management**: Organize users into groups for efficient alert routing
- **Web Dashboard**: React-based web interface for monitoring and managing alerts
- **Notification Providers**: Configurable priority-based notification provider system
- **Schedulers**: Background workers for processing alerts and messages
- **RESTful API**: Comprehensive REST API for integration
- **JWT Authentication**: Secure token-based authentication
- **CAPTCHA Support**: Built-in CAPTCHA for user registration

## Architecture

```
├── cmd/server/          # Application entry point
├── internal/            # Internal application code
│   ├── bootstrap/       # Application initialization
│   ├── config/          # Configuration management
│   ├── logging/         # Logging setup
│   ├── schedulers/      # Background schedulers
│   ├── server/          # HTTP server setup
│   └── storage/         # Database connections
├── pkg/                 # Reusable packages
│   ├── alerts/          # Alert management
│   ├── auth/            # Authentication
│   ├── groups/          # Group management
│   ├── http/            # HTTP handlers and middleware
│   ├── message/         # Message handling
│   ├── notifications/   # Notification providers
│   ├── roles/           # Role management
│   ├── storage/         # Data repositories
│   └── user/            # User management
├── migrations/          # Database migrations
└── web/                 # React frontend application
```

## Prerequisites

- **Go**: 1.23.0 or higher
- **PostgreSQL**: 13 or higher
- **Node.js**: 16 or higher (for web frontend)
- **Git**: For version control

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/lyralab/iris.git
cd iris
```

### 2. Install Dependencies

#### Backend (Go)
```bash
go mod download
```

#### Frontend (React)
```bash
cd web
npm install
cd ..
```

### 3. Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE iris;
CREATE USER iris_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE iris TO iris_user;
```

### 4. Run Database Migrations

The application automatically runs migrations on startup, or you can run them manually:

```bash
# Migrations are located in the migrations/ directory
# They will be executed automatically when the application starts
```

## Configuration

Iris uses environment variables for configuration. Create a `.env` file in the root directory with the following variables:

### Required Configuration

```bash
# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DATABASE_NAME=iris
POSTGRES_USER=iris_user
POSTGRES_PASS=your_password
POSTGRES_SSL=false

# HTTP Server Configuration
HTTP_PORT=9090

# Security
JWT_SECRET=your-secret-jwt-key-change-this-in-production
ADMIN_PASS=your-admin-password

# Application Mode
GO_ENV=debug  # Use "release" for production
SIGNUP_ENABLED=true  # Set to false to disable user registration
```

### Notification Providers (Optional)

#### Kavenegar (SMS Provider)
```bash
KAVENEGAR_API_TOKEN=your-kavenegar-api-token
KAVENEGAR_SENDER=your-sender-number
KAVENEGAR_ENABLED=true
KAVENEGAR_PRIORITY=1  # Lower number = higher priority
```

#### Smsir (SMS Provider)
```bash
SMSIR_API_TOKEN=your-smsir-api-token
SMSIR_LINE_NUMBER=your-smsir-line-number
SMSIR_ENABLED=false
SMSIR_PRIORITY=2
```

#### Email Configuration
```bash
EMAIL_HOST=smtp.example.com
EMAIL_PORT=587
EMAIL_USER=your-email@example.com
EMAIL_PASSWORD=your-email-password
EMAIL_FROM=noreply@example.com
EMAIL_ENABLED=false
```

### Scheduler Configuration (Optional)

```bash
# Mobile/SMS Scheduler
MOBILE_SCHEDULER_START_AT=1s
MOBILE_SCHEDULER_INTERVAL=600s
MOBILE_SCHEDULER_WORKERS=1
MOBILE_SCHEDULER_QUEUE_SIZE=1
MOBILE_SCHEDULER_CACHE_CAPACITY=1

# Alert Scheduler
ALERT_SCHEDULER_START_AT=2s
ALERT_SCHEDULER_INTERVAL=10s
ALERT_SCHEDULER_WORKERS=1
ALERT_SCHEDULER_QUEUE_SIZE=10

# Message Status Scheduler
MESSAGE_STATUS_START_AT=10s
MESSAGE_STATUS_INTERVAL=10s
MESSAGE_STATUS_WORKERS=10
MESSAGE_STATUS_QUEUE_SIZE=100
```

See `.env.example` for a complete configuration template.

## Running the Application

### Development Mode

#### Start the Backend
```bash
# From the project root
go run cmd/server/main.go
```

The API server will start on `http://localhost:9090` (or your configured HTTP_PORT).

#### Start the Frontend
```bash
# In a separate terminal
cd web
npm start
```

The web interface will be available at `http://localhost:3000`.

### Production Mode

#### Build the Backend
```bash
go build -o iris cmd/server/main.go
```

#### Run the Backend
```bash
./iris
```

#### Build the Frontend
```bash
cd web
npm run build
```

The production build will be created in the `web/build` directory. Serve it using a web server like Nginx or Apache.

## API Documentation

### Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### User Management

#### Sign Up
```http
POST /v1/users/create
Content-Type: application/json

{
  "username": "john.doe",
  "firstname": "John",
  "lastname": "Doe",
  "password": "SecurePass123!",
  "confirm-password": "SecurePass123!",
  "email": "john@example.com"
}
```

#### Sign In
```http
POST /v1/users/signin
Content-Type: application/json

{
  "username": "john.doe",
  "password": "SecurePass123!"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "username": "john.doe",
    "role": "viewer"
  }
}
```

#### Verify User (Admin only)
```http
POST /v1/users/verify
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "username": "john.doe"
}
```

### Alert Management

#### Get All Alerts
```http
GET /v0/alerts
Authorization: Bearer <token>
```

#### Get Alert by ID
```http
GET /v0/alerts/{alert_id}
Authorization: Bearer <token>
```

#### Get Alerts with Pagination
```http
GET /v0/alerts/?page=1&pagination=10&status=firing
Authorization: Bearer <token>
```

Status values: `firing`, `resolved`

#### Get Firing Alert Count
```http
GET /v0/alerts/firingCount
Authorization: Bearer <token>
```

### Webhook Endpoints

#### Grafana Webhook
```http
POST /v0/messages/grafana
Authorization: Basic admin:<ADMIN_PASS>
Content-Type: application/json

{
  "receiver": "My Super Webhook",
  "status": "firing",
  "alertService": [
    {
      "status": "firing",
      "labels": {
        "alertname": "High memory usage",
        "team": "backend",
        "receptor": "+1-555-0100",
        "method": "sms"
      },
      "annotations": {
        "description": "The system has high memory usage",
        "summary": "High memory alert"
      },
      "startsAt": "2024-01-12T09:51:03.157076+02:00",
      "fingerprint": "c6eadffa33fcdf37"
    }
  ]
}
```

#### AlertManager Webhook
```http
POST /v1/messages/alertmanager
Authorization: Basic admin:<ADMIN_PASS>
Content-Type: application/json

{
  "version": "4",
  "groupKey": "alert-group-1",
  "status": "firing",
  "receiver": "iris-webhook",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertName": "HighCPU",
        "method": "sms",
        "receptor": "admin",
        "severity": "critical"
      },
      "annotations": {
        "summary": "CPU usage is above 90%"
      },
      "startsAt": "2024-01-12T12:34:32.908Z",
      "fingerprint": "unique-fingerprint-123"
    }
  ]
}
```

### Health Check

```http
GET /health
```

Response:
```json
{
  "status": "healthy"
}
```

## Alert Label Configuration

For proper routing of notifications, include these labels in your alerts:

- **receptor**: The recipient's phone number or username (required)
- **method**: Notification method - `sms`, `call`, or `email` (required)
- **alertname**: Name of the alert
- **severity**: Alert severity level (e.g., `critical`, `warning`, `info`)

Example Grafana alert configuration:

```json
{
  "labels": {
    "alertname": "HighMemoryUsage",
    "receptor": "+1-555-0100",
    "method": "sms",
    "severity": "critical"
  },
  "annotations": {
    "summary": "Memory usage is above 80%",
    "description": "Server XYZ has memory usage of 85%"
  }
}
```

## Web Interface

The web dashboard provides a comprehensive admin interface for managing the Iris Alert System:

### Features

- **Admin Authentication**: Secure JWT-based authentication with captcha verification (admin users only)
- **Alert Dashboard**: Overview of firing and resolved alerts with severity breakdown
- **User Management**: Create, edit, verify, and manage system users
- **Group Management**: Create groups, view members, and manage group assignments
- **Role-Based Access Control**: All features restricted to admin users only

### Pages

1. **Login** (`/`): Admin-only authentication with captcha
2. **Alerts** (`/alerts`): Alert summary, firing issues, and resolved issues
3. **Users** (`/users`): User management interface
4. **Groups** (`/groups`): Group management and user assignment

### Configuration

The frontend is configured to connect to the backend API in `web/src/config.js`. Update the `base_url` to point to your API server:

```javascript
const base_url = 'http://127.0.0.1:9090';
```

### Access Control

- Only users with the `admin` role can access the web interface
- All pages except login require valid JWT authentication
- JWT tokens are stored in browser localStorage
- Tokens include user role information for authorization

### Using the Web Interface

#### First Time Setup

1. Start the backend server (ensure database migrations have run)
2. The default admin user is created automatically:
   - Username: `admin`
   - Password: As configured in `ADMIN_PASS` environment variable
3. Build and serve the frontend (see Running the Application section above)
4. Navigate to `http://localhost:3000`
5. Login with admin credentials and complete the captcha

#### Managing Users

1. Navigate to **Users** page from the navigation bar
2. Click **Create User** to add a new user
3. Fill in user details (username, name, email, password)
4. New users are created in `pending` status
5. Click **Verify** to activate pending users
6. Click **Edit** to modify user information

#### Managing Groups

1. Navigate to **Groups** page from the navigation bar
2. Click **Create Group** to add a new group
3. Enter group name and optional description
4. Click **View Members** on any group to see and manage members
5. Click **Add User** within the members view to assign users to the group
6. Select a user from the dropdown and submit

#### Viewing Alerts

1. Navigate to **Alerts** page to see the alert dashboard
2. View alert summary by severity (Critical, High, Medium, Low, Warning, Page)
3. See top 10 firing issues with filter by severity
4. View latest resolved issues with duration information

## Database Schema

The system uses the following main tables:

- **alerts**: Stores all alert information
- **users**: User accounts
- **roles**: User roles and permissions
- **groups**: User groups for organization
- **user_groups**: Many-to-many relationship between users and groups
- **providers**: Notification provider configuration
- **message**: Outgoing notification messages

## Security Considerations

1. **Change Default Passwords**: Always change the default admin password
2. **JWT Secret**: Use a strong, unique JWT secret in production
3. **HTTPS**: Use HTTPS in production environments
4. **Database Security**: Use strong database passwords and restrict access
5. **API Keys**: Keep notification provider API keys secure
6. **Disable Signup**: Set `SIGNUP_ENABLED=false` in production if you don't want open registration

## Troubleshooting

### Application Won't Start

- **Error: "JWT_SECRET not set"**
  - Solution: Set the `JWT_SECRET` environment variable

- **Error: Database connection failed**
  - Solution: Check database credentials and ensure PostgreSQL is running

### Notifications Not Being Sent

- Check that the appropriate notification provider is enabled
- Verify API tokens are correct
- Check scheduler configuration
- Review application logs for errors

### Web Interface Can't Connect to API

- Verify the backend is running
- Check the `base_url` in `web/src/config.js`
- Ensure CORS is properly configured

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/alerts/...
```

### Code Structure

- Follow Go best practices and conventions
- Use meaningful variable and function names
- Add comments for exported functions
- Write tests for new features

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is proprietary software. All rights reserved.

## Support

For issues and questions:
- Create an issue on GitHub
- Contact the development team

## Changelog

### Version 1.0.0
- Initial release
- Alert management system
- Multi-channel notifications
- User and role management
- Web dashboard
- Grafana and AlertManager integration
