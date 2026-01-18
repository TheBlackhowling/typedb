-- Create a comprehensive table demonstrating ALL SQL Server data types
CREATE TABLE type_examples (
    id INT IDENTITY(1,1) PRIMARY KEY,
    
    -- Exact numeric types
    tiny_int TINYINT,
    small_int SMALLINT,
    integer_col INT,
    big_int BIGINT,
    decimal_col DECIMAL(10, 2),
    numeric_col NUMERIC(10, 2),
    money_col MONEY,
    smallmoney_col SMALLMONEY,
    bit_col BIT,
    
    -- Approximate numeric types
    float_col FLOAT,
    real_col REAL,
    
    -- Character string types
    char_col CHAR(10),
    varchar_col VARCHAR(255),
    varchar_max_col VARCHAR(MAX),
    
    -- Unicode character string types
    nchar_col NCHAR(10),
    nvarchar_col NVARCHAR(255),
    nvarchar_max_col NVARCHAR(MAX),
    
    -- Legacy text types
    text_col TEXT,
    ntext_col NTEXT,
    
    -- Binary string types
    binary_col BINARY(16),
    varbinary_col VARBINARY(255),
    varbinary_max_col VARBINARY(MAX),
    image_col IMAGE,
    
    -- Date and time types
    date_col DATE,
    time_col TIME,
    datetime_col DATETIME,
    datetime2_col DATETIME2,
    datetimeoffset_col DATETIMEOFFSET,
    smalldatetime_col SMALLDATETIME,
    
    -- Special types
    timestamp_col TIMESTAMP,
    uniqueidentifier_col UNIQUEIDENTIFIER,
    xml_col XML,
    hierarchyid_col HIERARCHYID,
    geography_col GEOGRAPHY,
    geometry_col GEOMETRY,
    sql_variant_col SQL_VARIANT,
    
    -- Created timestamp
    created_at DATETIME2 DEFAULT GETDATE()
);
