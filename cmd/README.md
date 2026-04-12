# Commands

## migrate_fresh.go

Truncates all database tables for a fresh migration, similar to Laravel's `migrate:fresh` command.

### Usage

Run directly:
```bash
go run cmd/migrate_fresh.go
```

Or build as a standalone binary:
```bash
go build -o migrate_fresh cmd/migrate_fresh.go
./migrate_fresh
```

### What it does

- Disables foreign key constraints temporarily
- Truncates all tables in the correct order (respecting dependencies)
- Resets auto-increment sequences
- Re-enables foreign key constraints

### Tables truncated (in order)

1. `user_role_permissions` (junction table)
2. `users`
3. `user_roles`
4. `user_permissions`

### Note

This command will **DELETE ALL DATA** from your database. Use with caution in production environments.

After running this command, the database will be empty. The next time you start the application, the seeders in `config/db.go` will run automatically to create:
- Default user permissions
- Default user roles (superadmin, user)
- Default role permissions (superadmin gets all permissions)
- Default admin user (from .env)
