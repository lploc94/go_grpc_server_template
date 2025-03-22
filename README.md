# Go gRPC Server Template

A production-ready template for building gRPC servers in Go with built-in database connectivity, logging, and graceful shutdown handling.

## Features

- **Complete gRPC Server Implementation** - Ready to extend with your services
- **Database Integration** - Pre-configured TiDB/MySQL connectivity using GORM
- **Structured Logging** - File-based logging with rotation by date
- **Graceful Shutdown** - Proper signal handling and connection cleanup
- **Environment Configuration** - Using .env files with godotenv
- **Protocol Buffers** - Sample proto definition and pre-configured compilation
- **Well-Documented Code** - Extensive comments explaining each component

## Prerequisites

- Go 1.18 or later
- Protocol Buffer Compiler (protoc)
- Go plugins for Protocol Buffers:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

## Getting Started

### Installation

1. Clone this repository
   ```bash
   git clone https://github.com/lploc94/go_grpc_server_template.git
   cd go_grpc_server_template
   ```

2. Modify the proto file in protoc directory to define your service

3. Generate Go code from proto files
   ```bash
   protoc --proto_path=./protoc --go_out=. --go-grpc_out=. protoc/myservice.proto
   ```

4. Implement your service methods by modifying the existing code in main.go

5. Create a `.env` file based on the example below:
   ```
   GRPC_LISTEN_PORT=12345
   LOG_DIR=logs
   TIDB_HOST=localhost
   TIDB_PORT=4000
   TIDB_USER=root
   TIDB_DATABASE=test
   ```

6. Build and run the server
   ```bash
   go build -o my-grpc-server
   ./my-grpc-server
   ```

## Project Structure

```
.
├── main.go                 # Main application entry point
├── protoc/                 # Protocol buffer definitions
│   └── myservice.proto     # Sample service definition
├── logs/                   # Log files directory
├── test.env                # Environment configuration
└── README.md               # This file
```

## Customizing the Template

### 1. Define Your Own Service

Edit the myservice.proto file to define your service:

```protobuf
syntax = "proto3";

package yourservice;

option go_package = "protoc/yourservice";

message YourRequest {
    string field1 = 1;
    int32 field2 = 2;
}

message YourResponse {
    string result = 1;
}

service YourService {
    rpc YourMethod(YourRequest) returns (YourResponse);
}
```

Generate code from your proto file:

```bash
protoc --proto_path=./protoc --go_out=. --go-grpc_out=. protoc/yourservice.proto
```

### 2. Implement Your Service

Create a handler struct in main.go:

```go
type YourService struct {
    yourservice.UnimplementedYourServiceServer
    app *Application
}

func (s *YourService) YourMethod(ctx context.Context, req *yourservice.YourRequest) (*yourservice.YourResponse, error) {
    // Implement your logic here
    return &yourservice.YourResponse{Result: "success"}, nil
}
```

### 3. Register Your Service

Update the `setup` method:

```go
yourservice.RegisterYourServiceServer(
    app.server,
    &YourService{app: app},
)
```

## Database Usage

The template uses GORM with TiDB/MySQL. Define your data models and use them in your handlers:

```go
// Define your model
type YourModel struct {
    ID   string `gorm:"primaryKey"`
    Name string
}

// Use in your handler
func (s *YourService) YourMethod(ctx context.Context, req *yourservice.YourRequest) (*yourservice.YourResponse, error) {
    record := YourModel{ID: uuid.New().String(), Name: req.Field1}
    if err := s.app.tidbDatabase.Create(&record).Error; err != nil {
        return nil, status.Errorf(codes.Internal, "database error: %v", err)
    }
    return &yourservice.YourResponse{Result: "created"}, nil
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [gRPC-Go](https://github.com/grpc/grpc-go)
- [GORM](https://gorm.io/)
- [godotenv](https://github.com/joho/godotenv)