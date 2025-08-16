# Expenses Server

## Overview

The Expenses Server is a gRPC-based application designed to manage user data and expenses. It provides endpoints for creating users, retrieving user information, and uploading expense data (including image files) via streaming. The server is built using Go and follows a modular architecture with separate components for client, services, and database migrations.

## Features

- [x] **User Management**: Create and retrieve user information.
- [x] **Expense Management**: Upload expense data, including image files, using gRPC streaming.
- [x] **Database Migrations**: Manage database schema using `migrate`.
- [x] **GenAI data transcription** (using gemini API): Automatically transcribe expense data from uploaded images using the Gemini API for seamless integration.
- [ ] **Dashboard Integration**: Develop a user-friendly dashboard to visualize and manage expenses, including charts, summaries, and detailed views.
- [ ] **Local AI Model**: Integrate a local AI model for offline expense categorization and analysis, ensuring privacy and faster processing.
- [ ] **Personalised Trained Model**: Train and deploy personalized AI models for each user to provide tailored insights and recommendations based on spending habits.
- [ ] **Reads Expenses from Messages** (with Permission): Implement a feature to parse and extract expense data from user messages (e.g., SMS or emails) with explicit user consent.

## Prerequisites

Before running the application, ensure the following are installed:

1. **Go**: Install Go from [https://golang.org/dl/](https://golang.org/dl/).
2. **Protocol Buffers Compiler (protoc)**: Install `protoc` from [https://grpc.io/docs/protoc-installation/](https://grpc.io/docs/protoc-installation/).
3. **Migrate**: Install `migrate` for database migrations:
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```
4. **PostgreSQL**: Install PostgreSQL and ensure it is running.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/barathsurya2004/expenses-server.git
   cd expenses-server
   ```

2. Compile the Protocol Buffers:

   ```bash
   make idlCompile
   ```

3. Set up the database:

   - Create a PostgreSQL database.
   - Apply migrations:
     ```bash
     make migrateUp
     ```

4. Build and running the application:
   - Build the gRPC server:
     ```bash
     make serve
     ```
   - Build the client:
     ```bash
     make run
     ```

## Environment Variables

Before running the application, create a `.env` file in the root directory with the following template:

```env
# Database Configuration
POSTGRES_URL=postgres://<username>:<password>@<host>:<port>/<database>

# Server Configuration
GRPC_SERVER_PORT=50051
HTTP_SERVER_PORT=8080

# Other Configurations
CHUNK_SIZE=65536 # 64 KB
```

Replace `<username>`, `<password>`, `<host>`, `<port>`, and `<database>` with your PostgreSQL credentials and database details.

## Documentation

### Endpoints

#### 1. **Create User**

- **URL**: `/create-user`
- **Method**: `POST`
- **Description**: Creates a new user.

#### 2. **Get User**

- **URL**: `/get-user`
- **Method**: `GET`
- **Description**: Retrieves user information.

#### 3. **Create Expense**

- **URL**: `/create-expense`
- **Method**: `POST`
- **Description**: Uploads an expense image and streams it to the gRPC server.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
