// handler/report_handler.go — Handler with SQL queries (layer violation).
package handler

import (
	"database/sql"  // <-- VIOLATION: Handler importing database/sql
	"encoding/json"
	"net/http"
	"super-lint-test/pkg/httputil"
	"super-lint-test/pkg/logging"

	"github.com/gorilla/mux"
)

// ReportHandler generates reports with direct SQL.
type ReportHandler struct {
	db     *sql.DB  // <-- VIOLATION: Handler has database dependency
	logger logging.Logger
}

// NewReportHandler creates report handler with DB connection.
func NewReportHandler(db *sql.DB, logger logging.Logger) *ReportHandler {
	return &ReportHandler{
		db:     db,
		logger: logger,
	}
}

// GetUserReport handles GET /reports/users.
func (h *ReportHandler) GetUserReport(w http.ResponseWriter, r *http.Request) {
	// Handler directly executing SQL queries - massive architectural violation
	query := `
		SELECT
			DATE(created_at) as date,
			COUNT(*) as user_count,
			COUNT(CASE WHEN is_active THEN 1 END) as active_count
		FROM users
		WHERE created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`  // <-- VIOLATION: SQL query in handler layer

	rows, err := h.db.QueryContext(r.Context(), query)  // <-- VIOLATION: Direct DB access
	if err != nil {
		h.logger.Error("failed to query report", "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "failed to generate report")
		return
	}
	defer rows.Close()

	type ReportRow struct {
		Date        string `json:"date"`
		UserCount   int    `json:"user_count"`
		ActiveCount int    `json:"active_count"`
	}

	var results []ReportRow
	for rows.Next() {
		var row ReportRow
		if err := rows.Scan(&row.Date, &row.UserCount, &row.ActiveCount); err != nil {
			h.logger.Error("failed to scan row", "error", err)
			continue
		}
		results = append(results, row)
	}

	httputil.RespondJSON(w, http.StatusOK, results)
}
