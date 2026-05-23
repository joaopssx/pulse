CREATE TABLE IF NOT EXISTS incidents (
	id TEXT PRIMARY KEY,
	service_id TEXT NOT NULL,
	service_name TEXT NOT NULL,
	started_at DATETIME NOT NULL,
	resolved_at DATETIME,
	duration_ms INTEGER,
	cause TEXT,
	status_code INTEGER,
	latency_ms INTEGER,
	resolved INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS check_results (
	id TEXT PRIMARY KEY,
	service_id TEXT NOT NULL,
	service_name TEXT NOT NULL,
	checked_at DATETIME NOT NULL,
	status_code INTEGER,
	latency_ms INTEGER NOT NULL,
	success INTEGER NOT NULL,
	error TEXT
);
