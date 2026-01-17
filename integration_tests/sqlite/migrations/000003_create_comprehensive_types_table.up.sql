-- Create a comprehensive table demonstrating SQLite data types
-- Note: SQLite uses dynamic typing, but we can specify affinity types
CREATE TABLE IF NOT EXISTS type_examples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Numeric types (SQLite stores as INTEGER, REAL, or TEXT)
    integer_col INTEGER,
    real_col REAL,
    numeric_col NUMERIC(10, 2),
    
    -- Text types (SQLite stores as TEXT)
    text_col TEXT,
    varchar_col VARCHAR(255),
    char_col CHAR(10),
    clob_col CLOB,
    
    -- Blob types (SQLite stores as BLOB)
    blob_col BLOB,
    
    -- Date/Time types (stored as TEXT, INTEGER, or REAL)
    date_col DATE,
    datetime_col DATETIME,
    timestamp_col TIMESTAMP,
    time_col TIME,
    
    -- Boolean (stored as INTEGER: 0 or 1)
    boolean_col BOOLEAN,
    
    -- JSON (stored as TEXT, can use JSON functions)
    json_col TEXT, -- JSON stored as TEXT
    
    -- Created timestamp
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
