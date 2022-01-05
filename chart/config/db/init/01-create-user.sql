-- Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
-- Licensed under the MIT License. See License-MIT.txt in the project root for license information.

-- must be idempotent

-- create user (parameterized)
SET @statementStr = CONCAT(
    'CREATE USER IF NOT EXISTS "bhojpur"@"%" IDENTIFIED BY "', @bhojpurDbPassword, '";'
);
SELECT @statementStr ;
PREPARE stmt FROM @statementStr; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Grant privileges
GRANT ALL ON `bhojpur%`.* TO "bhojpur"@"%";
