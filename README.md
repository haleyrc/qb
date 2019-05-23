# qb

`qb` is a library for building simple SQL queries using a higher-level language suitable for reducing boilerplate for common operations.

The actual query building part does not offer 100% parity with the SQL spec, and there are a lot of much better and more battle-tested libraries out there for it. This is half learning experience and half start-of-some-pie-in-the-sky-idea-for-codegen.

**Disclaimer Time**

Per the above, if you use this in production, the onus is on you to make sure that it performs in the way you are expecting and if it doesn't, to help maintain the library by adding issues and PRs.

## TODO

- [ ] `AS` clauses for fields
- [ ] Extend paired ops (boolean and comparision) out to infinite number
- [X] `DELETE`
- [ ] `UPDATE`
- [ ] `INSERT`
- [ ] `LIMIT`
- [ ] `ORDER BY`

## Future

### Input

```
table users {
    id          uuid!;
    email       string!;
    first_name  string;
    last_name   string;
    admin       bool! = false;
}

query ListUsers {
    op = "select";
    table = "users";
    fields {
        id uuid;
        email string;
        first_name string;
        last_name string;
        admin bool;
    }
}

query ListAdmins {
    op = "select";
    table = "users";
    fields {
        id uuid;
        email string;
        first_name string;
        last_name string;
        admin bool;
    }
    where {
        admin = true;
    }
}
```

### Output:

`migration.sql`:

```sql
CREATE TABLE users (
    id UUID NOT NULL,
    email TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    admin BOOLEAN NOT NULL DEFAULT false
);

--- Suggested index from where queries: users.admin
```

`users.go`:

```go
type User {
    ID          string  `db:"id"`
    Email       string  `db:"email"`
    FirstName   string  `db:"first_name"`
    LastName    string  `db:"last_name"`
    Admin       bool    `db:"admin"`
}

const ListUsersQuery = `SELECT id, email, first_name, last_name, admin FROM users;`

const ListAdminsQuery = `SELECT id, email, first_name, last_name, admin FROM users WHERE admin = true;`

func ListUsers(db *sqlx.DB) ([]User, error) {
    var users []User
    if err := db.Select(&users, ListUsersQuery); err != nil {
        return nil, err
    }
    return users, nil
}

func ListAdmins(db *sqlx.DB) ([]User, error) {
    var users []User
    if err := db.Select(&users, ListAdminsQuery); err != nil {
        return nil err
    }
    return users, nil
}
```