package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	mzamqp "github.com/mnsrulz/mzworker-go/amqp"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start AMQP consumer daemon",
	Long:  "Start the AMQP consumer to process queries from a RabbitMQ queue.",
	RunE:  serveRunE,
}

func init() {
	serveCmd.Flags().IntP("concurrency", "c", 5, "Maximum concurrent message handlers")
	RootCmd.AddCommand(serveCmd)
}

func serveRunE(cmd *cobra.Command, args []string) error {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		return fmt.Errorf("DATA_DIR environment variable is required")
	}

	if err := registerMediatr(dataDir); err != nil {
		return err
	}

	amqpURL := os.Getenv("AMQP_URL")
	amqpQueue := os.Getenv("AMQP_REQUEST_QUEUE")
	if amqpURL == "" || amqpQueue == "" {
		return fmt.Errorf("AMQP_URL and AMQP_REQUEST_QUEUE environment variables are required")
	}

	concurrency, _ := cmd.Flags().GetInt("concurrency")

	consumer, err := mzamqp.NewConsumer(amqpURL, amqpQueue, concurrency)
	if err != nil {
		return fmt.Errorf("failed to create AMQP consumer: %w", err)
	}
	consumer.Start(cmd.Context())

	log.Println("AMQP consumer running. Press Ctrl+C to stop.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	consumer.Stop()
	return nil
}
