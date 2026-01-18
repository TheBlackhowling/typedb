-- Insert sample data demonstrating various SQLite types
INSERT INTO type_examples (
    integer_col, real_col, numeric_col,
    text_col, varchar_col, char_col, clob_col,
    blob_col,
    date_col, datetime_col, timestamp_col, time_col,
    boolean_col,
    json_col
) VALUES (
    2147483647, 3.14159, 1234.56,
    'TEXT example', 'VARCHAR example', 'CHAR      ', 'CLOB example with longer content',
    X'DEADBEEF',
    '2024-01-15', '2024-01-15 14:30:00', '2024-01-15 14:30:00', '14:30:00',
    1, -- TRUE
    '{"key": "value", "number": 42, "array": [1, 2, 3]}'
);
