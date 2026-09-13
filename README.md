# My New Project

A simple Go backend with PostgreSQL integration and an HTML frontend to submit user data.

## Setup

### 1. Install dependencies

### 2. Configure environment variables
Create a `.env` file in the project root with the following:
**Note:** `.env` is git-ignored and must be created manually on each machine.

### 3. Create the database table
In pgAdmin, run:
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100) UNIQUE
);
```

### 4. Run the server
Server starts at `http://localhost:8080`

### 5. Test it
Open `index.html` in a browser, fill in the form, and submit. Check the `users` table in pgAdmin to confirm the data was saved.
readme.md