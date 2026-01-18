-- Insert sample data demonstrating ALL MySQL types
INSERT INTO type_examples (
    tiny_int, tiny_int_unsigned, small_int, small_int_unsigned,
    medium_int, medium_int_unsigned, integer_col, integer_col_unsigned,
    big_int, big_int_unsigned,
    decimal_col, decimal_col_unsigned, numeric_col, numeric_col_unsigned,
    float_col, float_col_precision, double_col, double_col_precision,
    bit_col, bit_col_64,
    char_col, varchar_col, binary_col, varbinary_col,
    tinytext_col, text_col, mediumtext_col, longtext_col,
    tinyblob_col, blob_col, mediumblob_col, longblob_col,
    enum_col, set_col,
    date_col, time_col, datetime_col, timestamp_col, year_col,
    json_col,
    point_col
) VALUES (
    127, 255, 32767, 65535,
    8388607, 16777215, 2147483647, 4294967295,
    9223372036854775807, 18446744073709551615,
    1234.56, 1234.56, 7890.12, 7890.12,
    3.14159, 123.4567, 2.71828, 12345.67890123,
    b'10101010', b'1111000011110000111100001111000011110000111100001111000011110000',
    'CHAR      ', 'VARCHAR example', UNHEX('DEADBEEF'), UNHEX('CAFEBABE'),
    'TINYTEXT', 'TEXT example', 'MEDIUMTEXT example', 'LONGTEXT example with much longer content that exceeds normal text length',
    UNHEX('010203'), UNHEX('04050607'), UNHEX('08090A0B0C0D0E0F'), UNHEX('101112131415161718191A1B1C1D1E1F'),
    'value2', 'option1,option3',
    '2024-01-15', '14:30:00', '2024-01-15 14:30:00', '2024-01-15 14:30:00', 2024,
    '{"key": "value", "number": 42, "array": [1, 2, 3], "nested": {"data": true}}',
    ST_GeomFromText('POINT(10 20)')
);
