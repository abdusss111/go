-- Create expenses table
CREATE TABLE expenses (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency CHAR(3) NOT NULL,
    spent_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    note TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (category_id) REFERENCES categories(id),
    CHECK (amount > 0)
);

-- Create index on user_id for better query performance
CREATE INDEX idx_expenses_user_id ON expenses(user_id);

-- Create compound index on (user_id, spent_at) for optimized range queries
CREATE INDEX idx_expenses_user_spent_at ON expenses(user_id, spent_at);
