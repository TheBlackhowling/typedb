# Integration Test Coverage Audit

## Current Test Coverage Status

### PostgreSQL
**Tests Present:**
- ✅ QueryAll
- ✅ QueryFirst
- ✅ QueryOne
- ✅ Load
- ✅ LoadByField
- ✅ LoadByComposite
- ✅ Transaction (WithTx)
- ✅ PostgreSQLSpecificFeatures
- ✅ ComprehensiveTypes
- ✅ ComprehensiveTypesRoundTrip

**Tests Missing:**
- ❌ Insert (by object)
- ❌ InsertAndReturn
- ❌ InsertAndGetId
- ❌ Update (by object)
- ❌ QueryFirst error cases (no rows)
- ❌ QueryOne error cases (no rows, multiple rows)
- ❌ Negative test cases (invalid queries, constraint violations, etc.)

**Database Data Types Coverage:**
- ✅ SMALLINT, INTEGER, BIGINT
- ✅ DECIMAL, NUMERIC, REAL, DOUBLE PRECISION, MONEY
- ✅ CHAR, VARCHAR, TEXT
- ✅ BYTEA
- ✅ DATE, TIME, TIME WITH TIME ZONE, TIMESTAMP, TIMESTAMPTZ, INTERVAL
- ✅ BOOLEAN
- ✅ JSON, JSONB
- ✅ Arrays (all types)
- ✅ UUID
- ✅ INET, CIDR, MACADDR, MACADDR8
- ✅ Geometric types (POINT, LINE, LSEG, BOX, PATH, POLYGON, CIRCLE)
- ✅ Range types (INT4RANGE, INT8RANGE, NUMRANGE, TSRANGE, TSTZRANGE, DATERANGE)
- ✅ BIT, VARBIT
- ✅ TSVECTOR, TSQUERY
- ✅ XML

### MySQL
**Tests Present:**
- ✅ QueryAll
- ✅ QueryFirst
- ✅ QueryOne
- ✅ Load
- ✅ LoadByField
- ✅ LoadByComposite
- ✅ Transaction (WithTx)
- ✅ MySQLSpecificFeatures
- ✅ ComprehensiveTypes
- ✅ ComprehensiveTypesRoundTrip

**Tests Missing:**
- ❌ Insert (by object)
- ❌ InsertAndReturn
- ❌ InsertAndGetId
- ❌ Update (by object)
- ❌ QueryFirst error cases
- ❌ QueryOne error cases
- ❌ Negative test cases

**Database Data Types Coverage:**
- ✅ TINYINT, SMALLINT, MEDIUMINT, INT, BIGINT (signed and UNSIGNED)
- ✅ DECIMAL, NUMERIC (signed and UNSIGNED), FLOAT, DOUBLE
- ✅ BIT
- ✅ CHAR, VARCHAR, BINARY, VARBINARY
- ✅ TINYTEXT, TEXT, MEDIUMTEXT, LONGTEXT
- ✅ TINYBLOB, BLOB, MEDIUMBLOB, LONGBLOB
- ✅ ENUM, SET
- ✅ DATE, TIME, DATETIME, TIMESTAMP, YEAR
- ✅ JSON
- ✅ Geometry types

### SQLite
**Tests Present:**
- ✅ QueryAll
- ✅ QueryFirst
- ✅ QueryOne
- ✅ Load
- ✅ LoadByField
- ✅ LoadByComposite
- ✅ Transaction (WithTx)
- ✅ ComprehensiveTypes
- ✅ ComprehensiveTypesRoundTrip

**Tests Missing:**
- ❌ Insert (by object)
- ❌ InsertAndReturn
- ❌ InsertAndGetId
- ❌ Update (by object)
- ❌ QueryFirst error cases
- ❌ QueryOne error cases
- ❌ Negative test cases

**Database Data Types Coverage:**
- ✅ INTEGER, REAL, NUMERIC
- ✅ TEXT, VARCHAR, CHAR, CLOB
- ✅ BLOB
- ✅ DATE, DATETIME, TIMESTAMP, TIME
- ✅ BOOLEAN (as INTEGER)
- ✅ JSON (as TEXT)

### MSSQL
**Tests Present:**
- ✅ QueryAll
- ✅ Load
- ✅ LoadByField
- ✅ LoadByComposite
- ✅ Transaction (WithTx)
- ✅ ComprehensiveTypes
- ✅ ComprehensiveTypesRoundTrip

**Tests Missing:**
- ❌ QueryFirst
- ❌ QueryOne
- ❌ Insert (by object)
- ❌ InsertAndReturn
- ❌ InsertAndGetId
- ❌ Update (by object)
- ❌ Negative test cases

**Database Data Types Coverage:**
- ✅ TINYINT, SMALLINT, INT, BIGINT
- ✅ DECIMAL, NUMERIC, FLOAT, REAL, MONEY, SMALLMONEY
- ✅ BIT
- ✅ CHAR, VARCHAR, VARCHAR(MAX), NCHAR, NVARCHAR, NVARCHAR(MAX), TEXT, NTEXT
- ✅ BINARY, VARBINARY, VARBINARY(MAX), IMAGE
- ✅ DATE, TIME, DATETIME, DATETIME2, SMALLDATETIME, DATETIMEOFFSET, TIMESTAMP
- ✅ UNIQUEIDENTIFIER
- ✅ XML
- ✅ HIERARCHYID, GEOGRAPHY, GEOMETRY
- ✅ SQL_VARIANT

**Note:** MSSQL migration includes JSON but it's not in the model - need to verify if JSON type is tested

### Oracle
**Tests Present:**
- ✅ QueryAll
- ✅ Load
- ✅ LoadByField
- ✅ LoadByComposite
- ✅ Transaction (WithTx)
- ✅ ComprehensiveTypes
- ✅ ComprehensiveTypesRoundTrip
- ✅ LongRawType (separate table test)

**Tests Missing:**
- ❌ QueryFirst
- ❌ QueryOne
- ❌ Insert (by object)
- ❌ InsertAndReturn
- ❌ InsertAndGetId
- ❌ Update (by object)
- ❌ Negative test cases

**Database Data Types Coverage:**
- ✅ NUMBER, NUMBER(10,2), NUMBER(38)
- ✅ FLOAT, FLOAT(126)
- ✅ BINARY_FLOAT, BINARY_DOUBLE
- ✅ CHAR, VARCHAR2, VARCHAR, NCHAR, NVARCHAR2
- ✅ CLOB, NCLOB
- ✅ LONG (separate table)
- ✅ RAW, BLOB
- ✅ BFILE
- ✅ DATE, TIMESTAMP, TIMESTAMP(6), TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH LOCAL TIME ZONE
- ✅ INTERVAL YEAR TO MONTH, INTERVAL DAY TO SECOND
- ✅ ROWID, UROWID
- ✅ XMLTYPE

**Note:** Oracle migration mentions JSON but it's not in the model - need to verify if JSON type is tested

## Action Items

### 1. Add Missing API Tests (All Databases)
- [x] Insert (by object) tests ✅
- [x] InsertAndReturn tests ✅
- [x] InsertAndGetId tests ✅
- [x] Update (by object) tests ✅
- [x] QueryFirst error cases (no rows returns nil) ✅
- [x] QueryOne error cases (no rows = ErrNotFound, multiple rows = error) ✅
- [x] Negative test cases (invalid queries, constraint violations, etc.) ✅

### 2. Add Missing Tests for MSSQL and Oracle
- [x] QueryFirst tests ✅
- [x] QueryOne tests ✅
- [x] Insert, InsertAndReturn, InsertAndGetId, Update tests ✅
- [x] Negative tests ✅

### 3. Verify Database Data Type Coverage
- [x] All comprehensive type tests exist for all databases ✅
- [x] Round-trip tests verify deserialization ✅
- [ ] Verify all types from migrations are tested (need to audit)
- [ ] Ensure NULL handling is tested for nullable columns
- [ ] Test edge cases (zero values, max values, special values)

### 4. CI/CD Integration
- [ ] Verify all integration test workflows run correctly
- [x] Ensure tests fail (not skip) when databases unavailable ✅ (Required for CI/CD - databases must be available)
- [ ] Add test coverage reporting
