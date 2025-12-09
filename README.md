# Iris - Alert Notification System

Iris is a comprehensive alert management and notification system designed to receive alerts from monitoring systems like Grafana and AlertManager, and dispatch notifications through multiple channels including SMS and email.

## Features

- **Alert Management**: Receive, store, and manage alerts from AlertManager
- **Multi-Channel Notifications**: Support for SMS (Kavenegar, Smsir) 
- **User & Role Management**: Complete RBAC (Role-Based Access Control) system
- **Group Management**: Organize users into groups for efficient alert routing
- **Web Dashboard**: React-based web interface for monitoring and managing alerts
- **Notification Providers**: Configurable priority-based notification provider system
- **Schedulers**: Background workers for processing alerts and messages
- **RESTful API**: Comprehensive REST API for integration
- **JWT Authentication**: Secure token-based authentication
- **CAPTCHA Support**: Built-in CAPTCHA for user registration


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

The application automatically runs migrations on startup.

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

## Support

For issues and questions:
- Create an issue on GitHub
- Contact the development team

## License

* MIT
