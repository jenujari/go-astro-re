CREATE TABLE IF NOT EXISTS rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive', 'deprecated')),
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    current_version TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 100,
    is_modifier BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rule_versions (
    id BIGSERIAL PRIMARY KEY,
    rule_id TEXT NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('draft', 'active', 'inactive', 'deprecated')),
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    checksum TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (rule_id, version)
);

CREATE TABLE IF NOT EXISTS evaluation_requests (
    id BIGSERIAL PRIMARY KEY,
    request_datetime TIMESTAMPTZ NOT NULL,
    request_timezone TEXT NOT NULL,
    location_name TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    country_code TEXT NOT NULL,
    calculation_profile TEXT NOT NULL,
    request_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS astrology_context_snapshots (
    id BIGSERIAL PRIMARY KEY,
    evaluation_request_id BIGINT NOT NULL UNIQUE REFERENCES evaluation_requests(id) ON DELETE CASCADE,
    builder_version TEXT NOT NULL,
    context_payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS evaluation_results (
    id BIGSERIAL PRIMARY KEY,
    evaluation_request_id BIGINT NOT NULL UNIQUE REFERENCES evaluation_requests(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('success', 'partial_success', 'failed', 'canceled')),
    positive_total DOUBLE PRECISION NOT NULL,
    negative_total DOUBLE PRECISION NOT NULL,
    net_score DOUBLE PRECISION NOT NULL,
    partial_failure_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rule_results (
    id BIGSERIAL PRIMARY KEY,
    evaluation_result_id BIGINT NOT NULL REFERENCES evaluation_results(id) ON DELETE CASCADE,
    rule_id TEXT NOT NULL REFERENCES rules(id),
    rule_version TEXT NOT NULL,
    category TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('matched', 'not_matched', 'error')),
    matched BOOLEAN NOT NULL DEFAULT FALSE,
    polarity TEXT NOT NULL CHECK (polarity IN ('positive', 'negative', 'neutral')),
    raw_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    weight DOUBLE PRECISION NOT NULL DEFAULT 1,
    confidence_multiplier DOUBLE PRECISION NOT NULL DEFAULT 1,
    weighted_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    explanation TEXT NOT NULL,
    facts_used JSONB NOT NULL DEFAULT '[]'::jsonb,
    error_text TEXT NOT NULL DEFAULT '',
    duration_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rules_status ON rules(status);
CREATE INDEX IF NOT EXISTS idx_rules_category ON rules(category);
CREATE INDEX IF NOT EXISTS idx_rule_versions_rule_id ON rule_versions(rule_id);
CREATE INDEX IF NOT EXISTS idx_eval_requests_datetime ON evaluation_requests(request_datetime DESC);
CREATE INDEX IF NOT EXISTS idx_eval_results_status ON evaluation_results(status);
CREATE INDEX IF NOT EXISTS idx_rule_results_eval_result ON rule_results(evaluation_result_id);
CREATE INDEX IF NOT EXISTS idx_rule_results_rule_id ON rule_results(rule_id);
CREATE INDEX IF NOT EXISTS idx_rule_results_category ON rule_results(category);
