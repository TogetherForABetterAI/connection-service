-- Connection Service Database Schema
-- Table: client_sessions

-- Create the client_sessions table if it doesn't exist
CREATE TABLE IF NOT EXISTS client_sessions (
    session_id VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    session_status VARCHAR(50) NOT NULL CHECK (session_status IN ('IN_PROGRESS', 'COMPLETED', 'TIMEOUT')),
    dispatcher_status VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Create index on client_id for fast lookups
CREATE INDEX IF NOT EXISTS idx_client_sessions_client_id ON client_sessions(client_id);

-- Create index on session_status for filtering active sessions
CREATE INDEX IF NOT EXISTS idx_client_sessions_status ON client_sessions(session_status);

-- Create composite index for common query pattern (client_id + session_status)
CREATE INDEX IF NOT EXISTS idx_client_sessions_client_status ON client_sessions(client_id, session_status);

-- Create index on created_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_client_sessions_created_at ON client_sessions(created_at DESC);

-- Comments for documentation
COMMENT ON TABLE client_sessions IS 'Stores client session information for tracking connection state and progress';
COMMENT ON COLUMN client_sessions.session_id IS 'Unique identifier for the session (UUID)';
COMMENT ON COLUMN client_sessions.client_id IS 'Identifier for the client associated with this session';
COMMENT ON COLUMN client_sessions.session_status IS 'Current status of the session: IN_PROGRESS, COMPLETED, or TIMEOUT';
COMMENT ON COLUMN client_sessions.dispatcher_status IS 'Status of the data dispatcher service for this session';
COMMENT ON COLUMN client_sessions.created_at IS 'Timestamp when the session was created';
COMMENT ON COLUMN client_sessions.completed_at IS 'Timestamp when the session was completed';
