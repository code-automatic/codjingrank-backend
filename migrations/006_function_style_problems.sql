-- Migration to support function-style code execution problems
-- Adds new columns for driver code, initial code, and execution mode.

ALTER TABLE problems
ADD COLUMN IF NOT EXISTS execution_mode varchar(50) NOT NULL DEFAULT 'function',
ADD COLUMN IF NOT EXISTS function_name varchar(255),
ADD COLUMN IF NOT EXISTS initial_code text,
ADD COLUMN IF NOT EXISTS driver_code text,
ADD COLUMN IF NOT EXISTS solution_file_name varchar(255);
