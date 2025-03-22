package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/lploc94/go_grpc_server_template/protoc/myservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Application is the main application struct
// it contains the gRPC server, network listener, TiDB database, and log file.
// It also contains the setup, start, and stop methods for the application.
type Application struct {
	// Server is the gRPC server
	server *grpc.Server
	// NetListener is the network listener
	netListener net.Listener
	// tidbDatabase is the TiDB database
	tidbDatabase *gorm.DB
	// logFile is the log file
	logFile *os.File
}

// MyService is the gRPC service struct
// It contains a reference to the Application struct for accessing setup resources from the service methods.
type MyService struct {
	myservice.UnimplementedMyServiceServer
	app *Application
}

// TableRecord is a struct representing a record in the database table.
// It contains fields for the record columns.
// The struct tags define the column names and constraints for the GORM library.
// The A field is the primary key and unique index, while the B field is a regular column.
type TableRecord struct {
	A string `gorm:"column:a;primaryKey,uniqueIndex"`
	B int32  `gorm:"column:B"`
}

// setup method initializes the application by loading configuration from an environment file,
// setting up logging to a file, creating a gRPC server, and connecting to a TiDB database.
//
// Parameters:
//   - configPath: The path to the environment configuration file
//
// Returns:
//   - An error if the setup process fails
func (app *Application) setup(configPath string) error {
	err := godotenv.Load(configPath)
	if err != nil {
		return err
	}
	// Set up logging to file
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "logs"
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file with date in filename
	timestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("my-server-%s.log", timestamp))
	app.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Configure the logger to write to file and include timestamps
	log.SetOutput(app.logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// Create gRPC server
	app.server = grpc.NewServer()
	// Register the MyService server
	myservice.RegisterMyServiceServer(
		app.server,
		&MyService{app: app},
	)
	// Listen on the specified port
	app.netListener, err = net.Listen("tcp", ":"+os.Getenv("GRPC_LISTEN_PORT"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}
	// Connect to the TiDB database
	tidbConnectionString := os.Getenv("TIDB_USER") + ":@tcp(" + os.Getenv("TIDB_HOST") + ":" + os.Getenv("TIDB_PORT") + ")/" + os.Getenv("TIDB_DATABASE") + "?parseTime=true"
	app.tidbDatabase, err = gorm.Open(mysql.Open(tidbConnectionString), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to TiDB: %v", err)
	}
	return nil
}

// start method starts the gRPC server and listens for incoming requests.
func (app *Application) start() {
	log.Printf("Server listening on port %s", os.Getenv("GRPC_LISTEN_PORT"))
	if err := app.server.Serve(app.netListener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

// stop method stops the gRPC server gracefully by calling GracefulStop with a timeout.
func (app *Application) stop() {
	log.Println("Stopping server gracefully...")

	// Create a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use GracefulStop with deadline
	stopped := make(chan struct{})
	go func() {
		app.server.GracefulStop()

		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("Server stopped gracefully")
	case <-ctx.Done():
		log.Println("Force stopping server due to timeout")
		app.server.Stop()
	}
	// Close network listener

	if err := app.netListener.Close(); err != nil {
		log.Printf("Error closing listener: %v", err)
	}
	// Close database connection
	sqlDB, err := app.tidbDatabase.DB()
	if err == nil {
		sqlDB.Close()
		log.Println("Database connection closed")
	}
	// Close log file
	if app.logFile != nil {
		app.logFile.Close()
	}
	log.Println("Server shutdown complete")
}

func main() {
	app := Application{}
	err := app.setup("test.env")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set up signal handling first
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go app.start()

	// Wait for termination signal
	<-c
	app.stop()
}

// function MyMethod receives a request, creates a record in the database, and returns a response.
//
// Parameters:
//   - ctx: The context of the request
//   - req: The request message
//
// Returns:
//   - The response message
//   - An error if the operation failed
func (s *MyService) MyMethod(ctx context.Context, req *myservice.MyRequest) (*myservice.MyResponse, error) {
	// Perform some operation
	record := TableRecord{A: req.A, B: req.B}
	result := s.app.tidbDatabase.Create(&record)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to create record: %v", result.Error)
	}

	// Return response
	return &myservice.MyResponse{Message: "success"}, nil
}
