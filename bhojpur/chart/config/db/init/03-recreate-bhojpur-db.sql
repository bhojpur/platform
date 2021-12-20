-- Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
-- Licensed under the MIT License. See License-MIT.txt in the project root for license information.

-- must be idempotent

-- @bhojpurDB contains name of the DB the script manipulates, 'bhojpur' by default.
-- Prepend the script with "SET @bhojpurDB = '`<db-name>`'" if needed otherwise
SET @bhojpurDB = IFNULL(@bhojpurDB, '`bhojpur`');

SET @statementStr = CONCAT('DROP DATABASE IF EXISTS ', @bhojpurDB);
PREPARE statement FROM @statementStr;
EXECUTE statement;

SET @statementStr = CONCAT('CREATE DATABASE ', @bhojpurDB, ' CHARSET utf8mb4');
PREPARE statement FROM @statementStr;
EXECUTE statement;
