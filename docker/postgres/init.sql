-- ローカル開発専用。アプリは初期化用のsuperuserで接続しない。
CREATE ROLE goravel LOGIN PASSWORD 'goravel_local'
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION;
CREATE DATABASE go_backend_sample OWNER goravel;
CREATE DATABASE go_backend_sample_test OWNER goravel;
REVOKE ALL ON DATABASE go_backend_sample FROM PUBLIC;
REVOKE ALL ON DATABASE go_backend_sample_test FROM PUBLIC;

\connect go_backend_sample
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
ALTER DATABASE go_backend_sample SET timezone TO 'UTC';
ALTER DATABASE go_backend_sample SET statement_timeout TO '10s';
ALTER DATABASE go_backend_sample SET lock_timeout TO '3s';
ALTER DATABASE go_backend_sample SET idle_in_transaction_session_timeout TO '30s';

\connect go_backend_sample_test
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
ALTER DATABASE go_backend_sample_test SET timezone TO 'UTC';
ALTER DATABASE go_backend_sample_test SET statement_timeout TO '10s';
ALTER DATABASE go_backend_sample_test SET lock_timeout TO '3s';
ALTER DATABASE go_backend_sample_test SET idle_in_transaction_session_timeout TO '30s';
