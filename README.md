# Advanced Programming

## RecipeHub

> Description: Our service to discover recipes from other people to get new impressions

### Team Members:
- Karakuzov Bekbolat | SE-2217
- Erkinkyzy Bakyt | SE-2214
- Kalimatova Aisha | SE-2214

## First Page Look
![Frame 14](https://github.com/zbekxzz/advanced-programming/assets/129783111/1eaef1e2-bfdb-4614-8463-5c4927680f56)

## Table of Contents

- [Getting Started](#getting-started)
- [Tools Used](#tools-used)
- [Usage](#usage)
  - [Launching the Application](#launching-the-application)
- [API Endpoints](#api-endpoints)
- [Contributing](#contributing)
- [License](#license)

## Getting Started

To get started with this application, follow the steps below:

### Prerequisites

- Go (Golang) installed on your machine
- PostgreSQL database server

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/user-management-app.git
```

2. Navigate to the project directory:
```bash
cd user-management-app
```

3. Install Go dependencies:
```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/gorilla/mux
```

4. Install PostgreSQL:
> Make sure you have PostgreSQL installed on your machine. If not, download and install it from the official PostgreSQL website.

5. Configure PostgreSQL connection:
> Update the database connection details in the main.go file:
```go
dsn := "user=postgres password=Zbekxzz3 dbname=advanced sslmode=disable"
```

6. Create the necessary tables:
> Run the migration script to automatically create the required tables:
```bash
migrate -database "postgres://username:password@localhost:5432/your_database?sslmode=disable" -path db/migrations up
```

7. Run the application:
```bash
go run main.go
```

8. Open your web browser and go to http://localhost:8080 to access the application.

## Tools Used
> - Go (Golang)
> - Gorilla Mux (HTTP router and dispatcher)
> - GORM (Go Object Relational Mapper)
> - PostgreSQL (Database)

## API Endpoints
- GET /users/{id}: Get user by ID
- PUT /users/{id}: Update user's username by ID
- DELETE /users/{id}: Delete user by ID
- GET /users: Get a list of all users
- POST /api/register: Register a new user
- POST /api/login: Log in a user

## Contributing
Contributions are welcome! Feel free to open issues or submit pull requests.

## License
This project is licensed under the MIT License - see the LICENSE file for details.
