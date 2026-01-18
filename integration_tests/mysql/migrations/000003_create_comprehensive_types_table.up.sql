-- Create a comprehensive table demonstrating ALL MySQL data types
CREATE TABLE IF NOT EXISTS type_examples (
    id INT AUTO_INCREMENT PRIMARY KEY,
    
    -- Integer types
    tiny_int TINYINT,
    tiny_int_unsigned TINYINT UNSIGNED,
    small_int SMALLINT,
    small_int_unsigned SMALLINT UNSIGNED,
    medium_int MEDIUMINT,
    medium_int_unsigned MEDIUMINT UNSIGNED,
    integer_col INT,
    integer_col_unsigned INT UNSIGNED,
    big_int BIGINT,
    big_int_unsigned BIGINT UNSIGNED,
    
    -- Fixed-point types
    decimal_col DECIMAL(10, 2),
    decimal_col_unsigned DECIMAL(10, 2) UNSIGNED,
    numeric_col NUMERIC(10, 2),
    numeric_col_unsigned NUMERIC(10, 2) UNSIGNED,
    
    -- Floating-point types
    float_col FLOAT,
    float_col_precision FLOAT(7,4),
    double_col DOUBLE,
    double_col_precision DOUBLE(15,8),
    
    -- Bit type
    bit_col BIT(8),
    bit_col_64 BIT(64),
    
    -- Character types
    char_col CHAR(10),
    varchar_col VARCHAR(255),
    binary_col BINARY(16),
    varbinary_col VARBINARY(255),
    
    -- Text types
    tinytext_col TINYTEXT,
    text_col TEXT,
    mediumtext_col MEDIUMTEXT,
    longtext_col LONGTEXT,
    
    -- Blob types
    tinyblob_col TINYBLOB,
    blob_col BLOB,
    mediumblob_col MEDIUMBLOB,
    longblob_col LONGBLOB,
    
    -- Enum and Set types
    enum_col ENUM('value1', 'value2', 'value3', 'value4'),
    set_col SET('option1', 'option2', 'option3', 'option4'),
    
    -- Date/Time types
    date_col DATE,
    time_col TIME,
    datetime_col DATETIME,
    timestamp_col TIMESTAMP,
    year_col YEAR,
    
    -- JSON type
    json_col JSON,
    
    -- Geometry types
    geometry_col GEOMETRY,
    point_col POINT,
    linestring_col LINESTRING,
    polygon_col POLYGON,
    multipoint_col MULTIPOINT,
    multilinestring_col MULTILINESTRING,
    multipolygon_col MULTIPOLYGON,
    geometrycollection_col GEOMETRYCOLLECTION,
    
    -- Created timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
