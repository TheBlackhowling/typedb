-- Create a comprehensive table demonstrating ALL PostgreSQL data types
CREATE TABLE IF NOT EXISTS type_examples (
    id SERIAL PRIMARY KEY,
    
    -- Integer types
    small_int SMALLINT,
    integer_col INTEGER,
    big_int BIGINT,
    
    -- Arbitrary precision numeric
    decimal_col DECIMAL(10, 2),
    numeric_col NUMERIC(10, 2),
    
    -- Floating point
    real_col REAL,
    double_precision_col DOUBLE PRECISION,
    
    -- Monetary
    money_col MONEY,
    
    -- Character types
    varchar_col VARCHAR(255),
    char_col CHAR(10),
    text_col TEXT,
    
    -- Binary types
    bytea_col BYTEA,
    
    -- Date/Time types
    date_col DATE,
    time_col TIME,
    time_tz_col TIME WITH TIME ZONE,
    timestamp_col TIMESTAMP,
    timestamptz_col TIMESTAMPTZ,
    interval_col INTERVAL,
    
    -- Boolean
    boolean_col BOOLEAN,
    
    -- Enumerated type (requires CREATE TYPE)
    -- enum_col example_type_enum,
    
    -- JSON types
    json_col JSON,
    jsonb_col JSONB,
    
    -- Array types (all base types can be arrays)
    smallint_array SMALLINT[],
    int_array INTEGER[],
    bigint_array BIGINT[],
    real_array REAL[],
    double_precision_array DOUBLE PRECISION[],
    numeric_array NUMERIC(10,2)[],
    varchar_array VARCHAR(255)[],
    text_array TEXT[],
    boolean_array BOOLEAN[],
    date_array DATE[],
    timestamp_array TIMESTAMP[],
    json_array JSON[],
    jsonb_array JSONB[],
    uuid_array UUID[],
    bytea_array BYTEA[],
    
    -- UUID
    uuid_col UUID,
    
    -- Network address types
    inet_col INET,
    cidr_col CIDR,
    macaddr_col MACADDR,
    macaddr8_col MACADDR8,
    
    -- Geometric types
    point_col POINT,
    line_col LINE,
    lseg_col LSEG,
    box_col BOX,
    path_col PATH,
    polygon_col POLYGON,
    circle_col CIRCLE,
    
    -- Range types
    int4range_col INT4RANGE,
    int8range_col INT8RANGE,
    numrange_col NUMRANGE,
    tsrange_col TSRANGE,
    tstzrange_col TSTZRANGE,
    daterange_col DATERANGE,
    
    -- Bit string types
    bit_col BIT(8),
    varbit_col VARBIT(16),
    
    -- Text search types
    tsvector_col TSVECTOR,
    tsquery_col TSQUERY,
    
    -- XML
    xml_col XML,
    
    -- Composite type (stored as text/JSON representation)
    -- composite_col example_composite_type,
    
    -- Created timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
