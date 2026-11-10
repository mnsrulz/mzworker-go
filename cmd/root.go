package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/mnsrulz/mzworker-go/handler"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitSHA    = "unknown"
)

var RootCmd = &cobra.Command{
	Use:               "mzworker-go",
	Version:           Version,
	Short:             "CLI tool for querying DuckDB with options data",
	Long:              "CLI tool for querying DuckDB with SQL. Callable from opencode MCP/skills.",
	RunE:              rootRunE,
	PersistentPreRunE: configureLogging,
}

func init() {
	RootCmd.Flags().Bool("schema", false, "Show available tables and columns")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable debug logging (e.g. executed SQL)")
}

func configureLogging(cmd *cobra.Command, _ []string) error {
	level, ok := parseLogLevel(os.Getenv("LOG_LEVEL"))
	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
	if !ok {
		slog.Warn("invalid LOG_LEVEL; defaulting to info", "value", os.Getenv("LOG_LEVEL"))
	}
	return nil
}

func parseLogLevel(raw string) (slog.Level, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return slog.LevelInfo, true
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "warn", "warning":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	default:
		return slog.LevelInfo, false
	}
}

func rootRunE(cmd *cobra.Command, args []string) error {
	if schema, _ := cmd.Flags().GetBool("schema"); schema {
		printSchema()
		return nil
	}
	return cmd.Help()
}

func Execute() error {
	RootCmd.Version = Version
	return RootCmd.Execute()
}

func Output(columns []string, rows [][]interface{}, format string) error {
	switch format {
	case "table":
		return formatTable(os.Stdout, columns, rows)
	case "raw":
		return formatRaw(os.Stdout, columns, rows)
	default:
		return formatJSON(os.Stdout, columns, rows)
	}
}

func formatJSON(w *os.File, columns []string, rows [][]interface{}) error {
	objects := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		obj := make(map[string]interface{}, len(columns))
		for j, col := range columns {
			obj[col] = row[j]
		}
		objects[i] = obj
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(objects)
}

func formatTable(w *os.File, columns []string, rows [][]interface{}) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	for i, col := range columns {
		if i > 0 {
			_, _ = fmt.Fprint(tw, "\t")
		}
		_, _ = fmt.Fprint(tw, strings.ToUpper(col))
	}
	_, _ = fmt.Fprintln(tw)

	for _, row := range rows {
		for i, val := range row {
			if i > 0 {
				_, _ = fmt.Fprint(tw, "\t")
			}
			_, _ = fmt.Fprint(tw, formatValue(val))
		}
		_, _ = fmt.Fprintln(tw)
	}

	return tw.Flush()
}

func formatRaw(w *os.File, columns []string, rows [][]interface{}) error {
	type rawResponse struct {
		Columns []string        `json:"columns"`
		Rows    [][]interface{} `json:"rows"`
	}
	resp := rawResponse{Columns: columns, Rows: rows}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(resp)
}

func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	return fmt.Sprintf("%v", v)
}

func registerMediatr(dataDir string) error {
	return handler.Init(dataDir)
}
