CREATE TYPE order_status AS ENUM ('open', 'close');
CREATE TYPE payment_method AS ENUM ('cash', 'card', 'kaspi_qr');
CREATE TYPE item_size AS ENUM ('small', 'medium', 'large');
CREATE TYPE transaction_type AS ENUM ('added', 'written off', 'sale', 'created');

CREATE TABLE IF NOT EXISTS inventory (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    stock NUMERIC(10, 2) NOT NULL,
    price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    unit_type TEXT NOT NULL,
    last_updated TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    preferences JSONB
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    customer_id INT NOT NULL REFERENCES customers(id),
    total_amount NUMERIC(10, 2) NOT NULL CHECK (total_amount >= 0),
    status order_status NOT NULL DEFAULT 'open',
    special_instructions JSONB,
    payment_method payment_method NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS menu_items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    allergens TEXT[],
    size item_size NOT NULL,
    CONSTRAINT unique_menu_item_size UNIQUE (name, size)
);

CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id TEXT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    quantity NUMERIC NOT NULL CHECK (quantity > 0),
    price_at_order NUMERIC(10, 2) NOT NULL CHECK (price_at_order >= 0)
);

CREATE TABLE IF NOT EXISTS menu_item_ingredients (
    id SERIAL PRIMARY KEY,
    menu_item_id TEXT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    ingredient_id TEXT NOT NULL REFERENCES inventory(id) ON DELETE CASCADE,
    quantity NUMERIC NOT NULL CHECK (quantity > 0),
    CONSTRAINT unique_menu_item_ingredient UNIQUE (menu_item_id, ingredient_id)
);

CREATE TABLE IF NOT EXISTS order_status_history (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    previous_status order_status NOT NULL,
    new_status order_status NOT NULL,
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS price_history (
    id SERIAL PRIMARY KEY,
    menu_item_id TEXT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    old_price NUMERIC(10, 2) NOT NULL CHECK (old_price >= 0),
    new_price NUMERIC(10, 2) NOT NULL CHECK (new_price >= 0),
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory_transactions (
    id SERIAL PRIMARY KEY,
    inventory_id TEXT NOT NULL REFERENCES inventory(id) ON DELETE CASCADE,
    change_amount NUMERIC NOT NULL,
    transaction_type transaction_type NOT NULL,
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_menu_item_id ON order_items(menu_item_id);
CREATE INDEX idx_menu_items_name ON menu_items(name);
CREATE INDEX idx_inventory_name ON inventory(name);
CREATE INDEX idx_inventory_price ON inventory(price);
CREATE INDEX idx_inventory_stock_level ON inventory(stock);
CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX idx_price_history_menu_item_id ON price_history(menu_item_id);


INSERT INTO inventory (id, name, stock, unit_type, price)
VALUES
('coffee_beans', 'Coffee Beans', 100, 'kg', 15.0),
('milk', 'Milk', 50, 'liters', 1.5),
('sugar', 'Sugar', 200, 'kg', 0.8),
('flour', 'Flour', 150, 'kg', 0.5),
('butter', 'Butter', 80, 'kg', 5.0),
('chocolate', 'Chocolate', 60, 'kg', 10.0),
('vanilla_extract', 'Vanilla Extract', 30, 'liters', 25.0),
('eggs', 'Eggs', 120, 'units', 0.2),
('baking_powder', 'Baking Powder', 40, 'kg', 3.0),
('yeast', 'Yeast', 50, 'kg', 4.0),
('salt', 'Salt', 300, 'kg', 0.3),
('cocoa_powder', 'Cocoa Powder', 70, 'kg', 8.0),
('cheese', 'Cheese', 50, 'kg', 7.0),
('ham', 'Ham', 40, 'kg', 12.0),
('tomato', 'Tomato', 80, 'kg', 2.0),
('lettuce', 'Lettuce', 60, 'kg', 1.0),
('turkey', 'Turkey', 30, 'kg', 15.0),
('gluten', 'Gluten', 10, 'kg', 5.0),
('olive_oil', 'Olive Oil', 20, 'liters', 10.0),
('mayonnaise', 'Mayonnaise', 25, 'liters', 3.5),
('mustard', 'Mustard', 15, 'liters', 2.5);


INSERT INTO menu_items (id, name, description, price, allergens, size)
VALUES
('espresso', 'Espresso', 'Strong and bold coffee', 3.50, ARRAY['coffee'], 'small'),
('cappuccino', 'Cappuccino', 'Coffee with steamed milk foam', 4.50, ARRAY['coffee', 'milk'], 'medium'),
('latte', 'Latte', 'Coffee with steamed milk', 4.00, ARRAY['coffee', 'milk'], 'large'),
('americano', 'Americano', 'Espresso with hot water', 3.00, ARRAY['coffee'], 'medium'),
('flat_white', 'Flat White', 'Smooth coffee with microfoam', 4.20, ARRAY['coffee', 'milk'], 'small'),
('cheese_croissant', 'Cheese Croissant', 'Flaky pastry with cheese filling', 2.50, ARRAY['dairy'], 'medium'),
('chocolate_croissant', 'Chocolate Croissant', 'Flaky pastry with chocolate filling', 3.00, ARRAY['dairy', 'gluten'], 'medium'),
('muffin', 'Muffin', 'Freshly baked muffin', 2.80, ARRAY['gluten'], 'medium'),
('bagel', 'Bagel', 'Toasted bagel with cream cheese', 2.60, ARRAY['gluten', 'dairy'], 'medium');

INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity)
VALUES
('espresso', 'coffee_beans', 0.02), 
('cappuccino', 'coffee_beans', 0.02), 
('cappuccino', 'milk', 0.05),  
('latte', 'coffee_beans', 0.02), 
('latte', 'milk', 0.08),  
('americano', 'coffee_beans', 0.03), 
('flat_white', 'coffee_beans', 0.02), 
('flat_white', 'milk', 0.06), 
('cheese_croissant', 'flour', 0.1), 
('cheese_croissant', 'butter', 0.05), 
('cheese_croissant', 'cheese', 0.05), 
('chocolate_croissant', 'flour', 0.1), 
('chocolate_croissant', 'butter', 0.05), 
('chocolate_croissant', 'chocolate', 0.05), 
('muffin', 'flour', 0.1),  
('muffin', 'butter', 0.05), 
('muffin', 'sugar', 0.05),  
('bagel', 'flour', 0.12),  
('bagel', 'butter', 0.03),  
('bagel', 'cheese', 0.05),  
('sandwich', 'gluten', 0.15),  
('sandwich', 'cheese', 0.05),  
('sandwich', 'ham', 0.08);  

INSERT INTO customers (name, email, preferences)
VALUES
('Tauken Brave', 'john_smith@gmail.com', '{"note:":"subscribe_to_newsletters"}'),
('Emily Johnson', 'emily_johnson@gmail.com', '{"note:":"prefers_clothing_discounts"}'),
('Michael Williams', 'michael_williams@gmail.com', '{"note:":"not_interested_in_ads"}'),
('Sarah Brown', 'sarah_brown@gmail.com', '{"note:":"interested_in_electronics_promotions"}'),
('David Jones', 'david_jones@gmail.com', '{"note:":"wants_product_updates"}'),
('Olivia Garcia', 'olivia_garcia@gmail.com', '{"note:":"prefers_home_goods"}'),
('James Martinez', 'james_martinez@gmail.com', '{"note:":"interested_in_eco_friendly_products"}'),
('Sophia Rodriguez', 'sophia_rodriguez@gmail.com', '{"note:":"interested_in_new_books"}'),
('Daniel Wilson', 'daniel_wilson@gmail.com', '{"note:":"wants_travel_promotions"}'),
('Isabella Moore', 'isabella_moore@gmail.com', '{"note:":"prefers_cosmetics_discounts"}'),
('William Taylor', 'william_taylor@gmail.com', '{"note:":"interested_in_sports_and_fitness"}'),
('Charlotte Anderson', 'charlotte_anderson@gmail.com', '{"note:":"not_interested_in_newsletters"}'),
('Lucas Thomas', 'lucas_thomas@gmail.com', '{"note:":"interested_in_pets"}'),
('Mia Jackson', 'mia_jackson@gmail.com', '{"note:":"prefers_baby_products"}'),
('Henry White', 'henry_white@gmail.com', '{"note:":"looking_for_travel_deals"}');

INSERT INTO orders (customer_id, total_amount, status, special_instructions, payment_method, created_at, updated_at)
VALUES
(1, 7.00, 'open', '{"note": "No sugar"}', 'cash', '2024-01-15 10:00:00', '2024-01-15 10:05:00'),
(2, 5.50, 'close', '{"note": "Extra milk"}', 'card', '2024-02-16 14:30:00', '2024-02-16 14:35:00'),
(3, 4.20, 'open', '{"note": "No milk"}', 'kaspi_qr', '2024-03-17 08:10:00', '2024-03-17 08:12:00'),
(4, 6.00, 'close', '{"note": "Add extra shot"}', 'card', '2024-04-18 13:20:00', '2024-04-18 13:25:00'),
(5, 8.50, 'open', '{"note": "No cream"}', 'cash', '2024-05-19 09:30:00', '2024-05-19 09:35:00'),
(6, 5.00, 'close', '{"note": "Gluten-free"}', 'card', '2024-06-20 11:15:00', '2024-06-20 11:18:00'),
(7, 3.00, 'open', '{"note": "Extra cheese"}', 'cash', '2024-07-21 07:45:00', '2024-07-21 07:50:00'),
(8, 9.00, 'close', '{"note": "Spicy"}', 'kaspi_qr', '2024-08-22 15:30:00', '2024-08-22 15:35:00'),
(9, 6.80, 'open', '{"note": "No onions"}', 'cash', '2024-09-23 17:10:00', '2024-09-23 17:12:00'),
(10, 5.60, 'close', '{"note": "Light milk"}', 'card', '2024-10-24 12:25:00', '2024-10-24 12:30:00'),
(11, 7.20, 'open', '{"note": "Decaf"}', 'cash', '2024-11-25 08:40:00', '2024-11-25 08:42:00'),
(12, 4.50, 'close', '{"note": "No butter"}', 'kaspi_qr', '2024-12-26 10:15:00', '2024-12-26 10:20:00'),
(13, 6.30, 'open', '{"note": "Less sugar"}', 'card', '2024-01-27 13:45:00', '2024-01-27 13:50:00'),
(14, 5.10, 'close', '{"note": "Add nuts"}', 'cash', '2024-02-28 14:00:00', '2024-02-28 14:05:00'),
(1, 7.80, 'open', '{"note": "Double shot"}', 'card', '2024-03-29 16:25:00', '2024-03-29 16:30:00'),
(2, 8.00, 'open', '{"note": "No sugar, extra shot"}', 'cash', '2024-01-05 09:30:00', '2024-01-05 09:35:00'),
(3, 7.50, 'close', '{"note": "Extra milk, no butter"}', 'card', '2024-02-03 11:45:00', '2024-02-03 11:50:00'),
(4, 9.20, 'open', '{"note": "Gluten-free, no cheese"}', 'kaspi_qr', '2024-03-08 15:30:00', '2024-03-08 15:35:00'),
(5, 6.50, 'open', '{"note": "Double shot espresso"}', 'card', '2024-04-12 12:00:00', '2024-04-12 12:05:00'),
(6, 5.00, 'close', '{"note": "No onions, add cheese"}', 'cash', '2024-05-01 14:25:00', '2024-05-01 14:30:00'),
(7, 7.30, 'open', '{"note": "No cream, extra shot"}', 'card', '2024-06-15 16:40:00', '2024-06-15 16:45:00'),
(8, 4.80, 'close', '{"note": "Less sugar, extra foam"}', 'kaspi_qr', '2024-07-10 08:55:00', '2024-07-10 09:00:00'),
(9, 7.60, 'open', '{"note": "Add nuts, extra shot"}', 'cash', '2024-08-25 18:10:00', '2024-08-25 18:15:00'),
(10, 6.00, 'close', '{"note": "No milk, extra shot"}', 'card', '2024-09-12 14:05:00', '2024-09-12 14:10:00'),
(11, 5.90, 'open', '{"note": "Light milk, no butter"}', 'cash', '2024-10-19 13:25:00', '2024-10-19 13:30:00'),
(12, 8.40, 'close', '{"note": "Extra cheese, spicy"}', 'card', '2024-11-04 17:00:00', '2024-11-04 17:05:00'),
(10, 6.70, 'open', '{"note": "No butter, no sugar"}', 'kaspi_qr', '2024-12-06 10:15:00', '2024-12-06 10:20:00'),
(11, 5.20, 'close', '{"note": "Add extra shot"}', 'cash', '2024-01-09 11:10:00', '2024-01-09 11:15:00'),
(15, 6.40, 'open', '{"note": "Spicy, less milk"}', 'card', '2024-02-17 08:00:00', '2024-02-17 08:05:00'),
(3, 7.00, 'close', '{"note": "Double shot, no cream"}', 'kaspi_qr', '2024-03-19 19:30:00', '2024-03-19 19:35:00');


INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order)
VALUES
(1, 'espresso', 2, 3.50),  
(1, 'cheese_croissant', 1, 2.50),
(2, 'cappuccino', 1, 4.50),  
(2, 'muffin', 2, 2.80),
(3, 'latte', 1, 4.00),  
(3, 'chocolate_croissant', 1, 3.00),
(4, 'americano', 2, 3.00), 
(4, 'bagel', 1, 2.60),
(5, 'flat_white', 1, 4.20), 
(5, 'sandwich', 1, 5.50),
(6, 'cheese_croissant', 1, 2.50), 
(6, 'muffin', 2, 2.80),
(7, 'latte', 1, 4.00), 
(7, 'sandwich', 1, 5.50),
(8, 'americano', 1, 3.00),  
(8, 'muffin', 2, 2.80),
(9, 'espresso', 1, 3.50),  
(9, 'chocolate_croissant', 1, 3.00),
(10, 'cappuccino', 1, 4.50),  
(10, 'cheese_croissant', 1, 2.50),
(16, 'espresso', 2, 3.50),  
(16, 'cheese_croissant', 1, 2.50),
(17, 'latte', 1, 4.00),  
(17, 'muffin', 2, 2.80),
(18, 'americano', 1, 3.00), 
(18, 'bagel', 1, 2.60),
(19, 'flat_white', 1, 4.20), 
(19, 'sandwich', 1, 5.50),
(20, 'espresso', 1, 3.50),  
(20, 'cheese_croissant', 1, 2.50),
(21, 'cappuccino', 1, 4.50),  
(21, 'chocolate_croissant', 1, 3.00),
(22, 'latte', 1, 4.00),  
(22, 'muffin', 2, 2.80),
(23, 'espresso', 2, 3.50),
(23, 'sandwich', 1, 5.50),
(24, 'americano', 1, 3.00),  
(24, 'muffin', 2, 2.80),
(25, 'latte', 1, 4.00),  
(25, 'cheese_croissant', 1, 2.50),
(26, 'flat_white', 1, 4.20), 
(26, 'sandwich', 1, 5.50),
(27, 'cappuccino', 1, 4.50),  
(27, 'cheese_croissant', 1, 2.50),
(28, 'latte', 1, 4.00), 
(28, 'sandwich', 1, 5.50),
(29, 'espresso', 1, 3.50),  
(29, 'chocolate_croissant', 1, 3.00),
(30, 'cappuccino', 1, 4.50),  
(30, 'cheese_croissant', 1, 2.50);


INSERT INTO inventory_transactions (inventory_id, change_amount, transaction_type, changed_at)
VALUES
('coffee_beans', -0.04, 'sale', '2024-01-15 10:05:00'),
('cheese', -0.05, 'sale', '2024-02-16 14:35:00'),
('flour', -0.1, 'sale', '2024-03-17 08:12:00'),
('butter', -0.13, 'sale', '2024-04-18 13:25:00'),
('sugar', -0.05, 'sale', '2024-05-19 09:35:00'),
('chocolate', -0.05, 'sale', '2024-06-20 11:18:00'),
('cheese', -0.05, 'sale', '2024-07-21 07:50:00'),
('coffee_beans', -0.05, 'sale', '2024-08-22 15:35:00'),
('milk', -0.11, 'sale', '2024-09-23 17:12:00'),
('flour', -0.22, 'sale', '2024-10-24 12:30:00'),
('cheese', -0.1, 'sale', '2024-11-25 08:42:00'),
('ham', -0.16, 'sale', '2024-12-26 10:20:00'),
('flour', -0.1, 'sale', '2024-01-27 13:50:00'),
('cheese', -0.05, 'sale', '2024-02-28 14:05:00'),
('coffee_beans', -0.04, 'sale', '2024-01-05 09:35:00'),
('cheese', -0.05, 'sale', '2024-02-03 11:50:00'),
('flour', -0.1, 'sale', '2024-03-08 15:35:00'),
('butter', -0.13, 'sale', '2024-04-12 12:05:00'),
('sugar', -0.05, 'sale', '2024-05-01 14:30:00'),
('chocolate', -0.05, 'sale', '2024-06-15 16:45:00'),
('cheese', -0.05, 'sale', '2024-07-10 09:00:00'),
('coffee_beans', -0.05, 'sale', '2024-08-25 18:15:00'),
('milk', -0.11, 'sale', '2024-09-12 14:10:00'),
('flour', -0.22, 'sale', '2024-10-19 13:30:00'),
('cheese', -0.1, 'sale', '2024-11-04 17:05:00'),
('ham', -0.16, 'sale', '2024-12-06 10:20:00'),
('flour', -0.1, 'sale', '2024-01-09 11:15:00'),
('cheese', -0.05, 'sale', '2024-02-17 08:05:00'),
('coffee_beans', -0.04, 'sale', '2024-03-19 19:35:00');


INSERT INTO order_status_history (order_id, previous_status, new_status, changed_at)
VALUES
(2, 'open', 'close', '2024-12-02'),
(4, 'open', 'close', '2024-12-01'),
(6, 'open', 'close', '2024-12-02'),
(8, 'open', 'close', '2024-12-03'),
(10, 'open', 'close', '2024-12-02'),
(12, 'open', 'close', '2024-12-01'),
(14, 'open', 'close', '2024-12-01');


INSERT INTO price_history (menu_item_id, old_price, new_price, changed_at)
VALUES
('espresso', 3.00, 3.50, '2024-11-01'),
('cappuccino', 4.00, 4.50, '2024-11-01'),
('latte', 3.80, 4.00, '2024-11-01'),
('americano', 2.90, 3.00, '2024-11-01'),
('flat_white', 4.10, 4.20, '2024-11-01');
