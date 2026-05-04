CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    type INT NOT NULL,
    value INT NOT NULL,
    state VARCHAR(20) NOT NULL DEFAULT 'received',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_type ON tasks(type);
CREATE INDEX idx_tasks_state ON tasks(state);
