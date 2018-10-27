package stores

const schema = `
CREATE TABLE customers (
    id SERIAL,
    firstname VARCHAR(100) NOT NULL,
    lastname VARCHAR(100) NOT NULL,
    birthdate TIMESTAMP WITH TIME ZONE NOT NULL,
    gender VARCHAR(6) NOT NULL,
    email VARCHAR(254) NOT NULL,
    address VARCHAR(200) NOT NULL,

    CHECK (gender = 'Female' OR gender = 'Male')
);

CREATE INDEX customers_firstname_idx ON customers(firstname);
CREATE INDEX customers_lastname_idx ON customers(lastname);
CREATE INDEX customers_birthdate_idx ON customers(birthdate);
CREATE INDEX customers_gender_idx ON customers(gender);
CREATE INDEX customers_email_idx ON customers(email);
CREATE INDEX customers_address_idx ON customers(address);
`