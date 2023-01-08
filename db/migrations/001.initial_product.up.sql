CREATE TABLE product (
    id uuid PRIMARY KEY NOT NULL,
    sku text NOT NULL,
    name text NOT NULL,
    description text,
    price numeric NOT NULL
);
