# Database Support for Container Management

The container management system supports both SQLite and PostgreSQL databases for membership indexing and container metadata storage.

## Supported Databases

### SQLite (Default)
- **Driver**: `sqlite3`
- **Use Case**: Development, testing, single-user deployments
- **Advantages**: Zero configuration, embedded, file-based
- **Limitations**: Single writer, limited concurrency

### PostgreSQL
- **Driver**: `postgres`
- **Use Case**: Production deployments, multi-user systems
- **Advantages**: High concurrency, ACID compliance, scalability
- **Requirements**: PostgreSQL server installation

## Configuration

### Environment Variables

Set the following environment variables to configure the database:

```bash
# Database type
DB_DRIVER=postgres  # or sqlite3 (default)

# PostgreSQL configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goro_ldp
DB_USER=goro_user
DB_PASSWORD=your_password
DB_SSLMODE=disable  # or require for production

# SQLite configuration (when DB_DRIVER=sqlite3)
DB_PATH=./data/membership.db
```

### Programmatic Configuration

#### SQLite Configuration
```go
config := DefaultSQLiteConfig("./data/membership.db")
indexer, err := NewGenericMembershipIndexer(config)
```

#### PostgreSQL Configuration
```go
config := DefaultPostgresConfig("localhost", "goro_ldp", "user", "password")
indexer, err := NewGenericMembershipIndexer(config)
```

#### Environment-based Configuration
```go
config := DatabaseConfigFromEnv()
indexer, err := NewGenericMembershipIndexer(config)
```

## Database Schema

Both databases use the same logical schema with database-specific SQL syntax:

### Tables

#### containers
- `id` (TEXT/VARCHAR) - Primary key, container identifier
- `parent_id` (TEXT/VARCHAR) - Foreign key to parent container
- `type` (TEXT/VARCHAR) - Container type (BasicContainer, etc.)
- `title` (TEXT/VARCHAR) - Human-readable title
- `description` (TEXT/VARCHAR) - Container description
- `created_at` (TIMESTAMP) - Creation timestamp
- `updated_at` (TIMESTAMP) - Last modification timestamp

#### memberships
- `container_id` (TEXT/VARCHAR) - Foreign key to containers table
- `member_id` (TEXT/VARCHAR) - Resource or container ID
- `member_type` (TEXT/VARCHAR) - 'Container' or 'Resource'
- `created_at` (TIMESTAMP) - Membership creation timestamp
- Primary key: `(container_id, member_id)`

#### schema_migrations
- `version` (INTEGER) - Migration version number
- `description` (TEXT/VARCHAR) - Migration description
- `applied_at` (TIMESTAMP) - When migration was applied

### Indexes

Performance indexes are created for both databases:
- `idx_containers_parent` - On `containers(parent_id)`
- `idx_memberships_container` - On `memberships(container_id)`
- `idx_memberships_member` - On `memberships(member_id)`

## Migration System

The system includes automatic database migrations:

1. **Schema Versioning**: Tracks applied migrations in `schema_migrations` table
2. **Idempotent**: Safe to run multiple times
3. **Database-Agnostic**: Works with both SQLite and PostgreSQL

### Running Migrations

Migrations run automatically when creating a new indexer:

```go
indexer, err := NewGenericMembershipIndexer(config)
// Migrations are applied automatically
```

Manual migration:
```go
db, err := NewDatabaseConnection(config)
provider, err := NewSchemaProvider(config.Driver)
err = MigrateDatabaseWithProvider(db, provider)
```

## Performance Considerations

### SQLite
- **Concurrent Reads**: Excellent
- **Concurrent Writes**: Limited (single writer)
- **File Locking**: Uses filesystem locks
- **Memory Usage**: Low
- **Best For**: < 1000 containers, single-user scenarios

### PostgreSQL
- **Concurrent Reads**: Excellent
- **Concurrent Writes**: Excellent (MVCC)
- **Connection Pooling**: Recommended for high load
- **Memory Usage**: Configurable
- **Best For**: > 1000 containers, multi-user scenarios

## Production Deployment

### PostgreSQL Setup

1. **Install PostgreSQL**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install postgresql postgresql-contrib
   
   # macOS
   brew install postgresql
   ```

2. **Create Database and User**:
   ```sql
   CREATE DATABASE goro_ldp;
   CREATE USER goro_user WITH PASSWORD 'secure_password';
   GRANT ALL PRIVILEGES ON DATABASE goro_ldp TO goro_user;
   ```

3. **Configure Connection**:
   ```bash
   export DB_DRIVER=postgres
   export DB_HOST=localhost
   export DB_NAME=goro_ldp
   export DB_USER=goro_user
   export DB_PASSWORD=secure_password
   export DB_SSLMODE=require
   ```

### Security Considerations

- **SSL/TLS**: Use `sslmode=require` in production
- **Credentials**: Store passwords in environment variables or secrets management
- **Network**: Restrict database access to application servers only
- **Backups**: Implement regular database backups

### Monitoring

Monitor these metrics for database health:
- Connection count and pool utilization
- Query execution time
- Index usage and table scans
- Database size and growth rate
- Replication lag (if using replicas)

## Testing

### Unit Tests

Both database types are tested:
```bash
# Test SQLite implementation
go test -run TestGenericMembershipIndexer_SQLite

# Test PostgreSQL implementation (requires running PostgreSQL)
go test -run TestGenericMembershipIndexer_PostgreSQL
```

### Integration Tests

Test with real databases:
```bash
# Set up test PostgreSQL database
createdb test_goro_ldp
export DB_DRIVER=postgres
export DB_NAME=test_goro_ldp

# Run integration tests
go test ./internal/ldp/infrastructure/
```

## Troubleshooting

### Common Issues

1. **Connection Refused**:
   - Check database server is running
   - Verify host and port configuration
   - Check firewall settings

2. **Authentication Failed**:
   - Verify username and password
   - Check user permissions
   - Ensure database exists

3. **Migration Errors**:
   - Check database permissions
   - Verify schema_migrations table access
   - Review migration logs

4. **Performance Issues**:
   - Check index usage with EXPLAIN
   - Monitor connection pool settings
   - Review query patterns

### Debugging

Enable database query logging:
```go
// Add to development configuration
config.LogLevel = "debug"
```

Check migration status:
```sql
SELECT * FROM schema_migrations ORDER BY version;
```

## Future Enhancements

Planned database features:
- Connection pooling configuration
- Read replica support
- Database sharding for large deployments
- Automated backup integration
- Performance monitoring integration