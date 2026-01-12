# Seed Package

This package provides database seeding functionality for typedb examples. It uses [faker](https://github.com/jaswdr/faker) to generate randomized data for testing and demonstration purposes.

## Usage

Import the seed package in your example:

```go
import "github.com/TheBlackHowling/typedb/examples/seed"
```

## Functions

### SeedDatabase

Seeds the database with random data:

```go
err := seed.SeedDatabase(ctx, db, 10) // Seeds 10 users
```

This function will:
- Create `numUsers` users with random names and emails
- Create one profile per user with random bio, avatar URL, location, and website
- Create 2-5 posts per user with random titles, content, and published status

### ClearDatabase

Clears all seeded data:

```go
err := seed.ClearDatabase(ctx, db)
```

This function deletes all data from `posts`, `profiles`, and `users` tables in that order (respecting foreign key constraints).

## Schema

The seed package expects the following tables:

- **users** - `id`, `name`, `email`, `created_at`
- **profiles** - `id`, `user_id` (FK), `bio`, `avatar_url`, `location`, `website`, `created_at`
- **posts** - `id`, `user_id` (FK), `title`, `content`, `published`, `created_at`

## Models

The seed package defines models (`User`, `Profile`, `Post`) that match the expected schema. These models are registered with typedb and can be used in examples.

## Dependencies

- `github.com/TheBlackHowling/typedb` - The typedb package
- `github.com/jaswdr/faker` - For generating fake data
