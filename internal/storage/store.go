package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pulse/pulse/internal/checker"
	"github.com/pulse/pulse/internal/models"
)

//go:embed migrations/001_init.sql
var initSQL string

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(initSQL); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) SaveCheckResult(ctx context.Context, result checker.Result) error {
	if result.ID == "" {
		result.ID = generateID()
	}

	query := `INSERT INTO check_results (id, service_id, service_name, checked_at, status_code, latency_ms, success, error)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		result.ID,
		result.ServiceID,
		result.ServiceName,
		result.CheckedAt,
		result.StatusCode,
		result.LatencyMs,
		result.Success,
		result.Error,
	)
	return err
}

func (s *Store) SaveIncident(ctx context.Context, incident models.Incident) error {
	if incident.ID == "" {
		incident.ID = generateID()
	}

	var durationMs *int64
	if incident.Duration != nil {
		val := incident.Duration.Milliseconds()
		durationMs = &val
	}

	query := `INSERT INTO incidents (id, service_id, service_name, started_at, resolved_at, duration_ms, cause, status_code, latency_ms, resolved)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		incident.ID,
		incident.ServiceID,
		incident.ServiceName,
		incident.StartedAt,
		incident.ResolvedAt,
		durationMs,
		incident.Cause,
		incident.StatusCode,
		incident.LatencyMs,
		incident.Resolved,
	)
	return err
}

func (s *Store) ResolveIncident(ctx context.Context, incidentID string, resolvedAt time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var startedAt time.Time
	err = tx.QueryRowContext(ctx, "SELECT started_at FROM incidents WHERE id = ?", incidentID).Scan(&startedAt)
	if err != nil {
		return err
	}

	durationMs := resolvedAt.Sub(startedAt).Milliseconds()

	query := `UPDATE incidents SET resolved_at = ?, duration_ms = ?, resolved = 1 WHERE id = ?`
	_, err = tx.ExecContext(ctx, query, resolvedAt, durationMs, incidentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) GetActiveIncidents(ctx context.Context) ([]models.Incident, error) {
	query := `SELECT id, service_id, service_name, started_at, resolved_at, duration_ms, cause, status_code, latency_ms, resolved
	FROM incidents WHERE resolved = 0`
	return s.queryIncidents(ctx, query)
}

func (s *Store) GetIncidentsByService(ctx context.Context, serviceID string, limit int) ([]models.Incident, error) {
	query := `SELECT id, service_id, service_name, started_at, resolved_at, duration_ms, cause, status_code, latency_ms, resolved
	FROM incidents WHERE service_id = ? ORDER BY started_at DESC LIMIT ?`
	return s.queryIncidents(ctx, query, serviceID, limit)
}

func (s *Store) queryIncidents(ctx context.Context, query string, args ...any) ([]models.Incident, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		var durationMs sql.NullInt64
		var cause sql.NullString
		var statusCode sql.NullInt64
		var latencyMs sql.NullInt64

		err := rows.Scan(
			&inc.ID,
			&inc.ServiceID,
			&inc.ServiceName,
			&inc.StartedAt,
			&inc.ResolvedAt,
			&durationMs,
			&cause,
			&statusCode,
			&latencyMs,
			&inc.Resolved,
		)
		if err != nil {
			return nil, err
		}
		if durationMs.Valid {
			d := time.Duration(durationMs.Int64) * time.Millisecond
			inc.Duration = &d
		}
		if cause.Valid {
			inc.Cause = cause.String
		}
		if statusCode.Valid {
			inc.StatusCode = int(statusCode.Int64)
		}
		if latencyMs.Valid {
			inc.LatencyMs = latencyMs.Int64
		}
		incidents = append(incidents, inc)
	}
	return incidents, rows.Err()
}

func (s *Store) GetRecentResults(ctx context.Context, serviceID string, limit int) ([]checker.Result, error) {
	query := `SELECT id, service_id, service_name, checked_at, status_code, latency_ms, success, error
	FROM check_results WHERE service_id = ? ORDER BY checked_at DESC LIMIT ?`

	rows, err := s.db.QueryContext(ctx, query, serviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []checker.Result
	for rows.Next() {
		var r checker.Result
		var nsError sql.NullString
		var nsStatusCode sql.NullInt64

		err := rows.Scan(
			&r.ID,
			&r.ServiceID,
			&r.ServiceName,
			&r.CheckedAt,
			&nsStatusCode,
			&r.LatencyMs,
			&r.Success,
			&nsError,
		)
		if err != nil {
			return nil, err
		}
		if nsStatusCode.Valid {
			r.StatusCode = int(nsStatusCode.Int64)
		}
		if nsError.Valid {
			r.Error = nsError.String
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *Store) GetUptimePercent(ctx context.Context, serviceID string, since time.Time) (float64, error) {
	query := `SELECT COUNT(*), SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END)
	FROM check_results WHERE service_id = ? AND checked_at >= ?`

	var total int
	var successful sql.NullInt64
	err := s.db.QueryRowContext(ctx, query, serviceID, since).Scan(&total, &successful)
	if err != nil {
		return 0, err
	}

	if total == 0 {
		return 100.0, nil
	}

	return float64(successful.Int64) / float64(total) * 100.0, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
