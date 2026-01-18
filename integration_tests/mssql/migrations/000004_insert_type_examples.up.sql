-- Insert sample data demonstrating ALL SQL Server types
SET IDENTITY_INSERT type_examples ON;
INSERT INTO type_examples (
    id,
    tiny_int, small_int, integer_col, big_int,
    decimal_col, numeric_col, money_col, smallmoney_col, bit_col,
    float_col, real_col,
    char_col, varchar_col, varchar_max_col,
    nchar_col, nvarchar_col, nvarchar_max_col,
    text_col, ntext_col,
    binary_col, varbinary_col, varbinary_max_col, image_col,
    date_col, time_col, datetime_col, datetime2_col, datetimeoffset_col, smalldatetime_col,
    uniqueidentifier_col,
    xml_col,
    hierarchyid_col,
    geography_col,
    geometry_col,
    sql_variant_col
) VALUES (
    1,
    255, 32767, 2147483647, 9223372036854775807,
    1234.56, 7890.12, $100.50, $50.25, 1,
    3.14159, 2.71828,
    'CHAR      ', 'VARCHAR example', 'VARCHAR(MAX) example with very long content',
    N'NCHAR     ', N'NVARCHAR example', N'NVARCHAR(MAX) example with Unicode: 你好世界',
    'TEXT example', N'NTEXT example',
    0xDEADBEEF, 0xCAFEBABE, 0x0102030405060708, 0x090A0B0C0D0E0F10,
    '2024-01-15', '14:30:00', '2024-01-15 14:30:00', '2024-01-15 14:30:00.1234567', '2024-01-15 14:30:00 +00:00', '2024-01-15 14:30:00',
    NEWID(),
    '<root><element>value</element></root>',
    '/1/', -- HierarchyID example
    geography::STGeomFromText('POINT(-122.4194 37.7749)', 4326), -- San Francisco coordinates
    geometry::STGeomFromText('POINT(10 20)', 0),
    CAST('SQL_VARIANT example' AS SQL_VARIANT)
);
SET IDENTITY_INSERT type_examples OFF;
