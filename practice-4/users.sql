-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    balance DECIMAL(10,2) DEFAULT 0.00
);

-- Insert sample data
INSERT INTO users (name, email, balance) VALUES
('John Doe', 'john.doe@example.com', 1000.00),
('Jane Smith', 'jane.smith@example.com', 1500.00),
('Bob Johnson', 'bob.johnson@example.com', 750.00),
('Alice Brown', 'alice.brown@example.com', 2000.00)
ON CONFLICT (email) DO NOTHING;
