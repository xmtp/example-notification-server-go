SET statement_timeout = 0;

--bun:split

CREATE TABLE installations (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

--bun:split

CREATE TABLE device_delivery_mechanisms (
    id SERIAL PRIMARY KEY,
    installation_id TEXT REFERENCES installations(id) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    kind TEXT NOT NULL,
    token TEXT NOT NULL,
    UNIQUE (installation_id, kind, token)
)

--bun:split

CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    installation_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    topic TEXT NOT NULL,
    is_active BOOLEAN
)

--bun:split

CREATE INDEX CONCURRENTLY subscriptions_topic_is_active_idx ON public.subscriptions (topic, is_active);

--bun:split

CREATE INDEX CONCURRENTLY device_delivery_mechanisms_installation_id_idx ON public.device_delivery_mechanisms (installation_id);