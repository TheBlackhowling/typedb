-- Insert sample data demonstrating ALL Oracle types
INSERT INTO type_examples (
    number_col, number_precision_col, number_scale_col, float_col, float_precision_col,
    binary_float_col, binary_double_col,
    char_col, varchar2_col, varchar_col, nchar_col, nvarchar2_col,
    clob_col, nclob_col, long_col,
    raw_col,
    blob_col,
    bfile_col, -- BFILE requires directory object setup, will fail without proper setup
    date_col, timestamp_col, timestamp_precision_col, timestamp_tz_col, timestamp_ltz_col,
    interval_year_col, interval_day_col,
    urowid_col,
    xmltype_col
) VALUES (
    2147483647, 1234.56, 99999999999999999999999999999999999999, 3.14159, 2.71828,
    2.71828, 1.41421,
    'CHAR      ', 'VARCHAR2 example', 'VARCHAR example', N'NCHAR     ', N'NVARCHAR2 example',
    'CLOB example with longer content', N'NCLOB example', 'LONG example content',
    HEXTORAW('DEADBEEF'),
    UTL_RAW.CAST_TO_RAW('BLOB example'),
    NULL, -- BFILE requires directory object setup - will need special handling in library
    DATE '2024-01-15', TIMESTAMP '2024-01-15 14:30:00', TIMESTAMP '2024-01-15 14:30:00.123456',
    TIMESTAMP '2024-01-15 14:30:00 +00:00', TIMESTAMP '2024-01-15 14:30:00',
    INTERVAL '1-2' YEAR TO MONTH, INTERVAL '1 2:30:45.123456' DAY TO SECOND,
    NULL, -- urowid_col (UROWID is typically system-generated or from another table's ROWID)
    XMLTYPE('<root><element>value</element></root>')
);

COMMIT;
