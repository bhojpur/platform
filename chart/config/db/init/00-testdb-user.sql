-- Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
-- Licensed under the MIT License. See License-MIT.txt in the project root for license information.

-- create test DB user
SET @bhojpurDbPassword = IFNULL(@bhojpurDbPassword, 'test');
