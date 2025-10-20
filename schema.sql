DROP TABLE IF EXISTS customer, room, booking, review, hotel_service, service_request;
DROP TYPE IF EXISTS room_types, hotel_services;

-- Creazione tabelle
CREATE TYPE room_types AS ENUM ('basic', 'suite');
CREATE TYPE hotel_services AS ENUM ('cleaning', 'room_service', 'massage');

CREATE TABLE customer(
    id int generated always as identity primary key,
    cf varchar(16),
    customer_name varchar(16),
    age int check (age > 0),
    email varchar(30)
);

CREATE TABLE room(
    id int generated always as identity primary key,
    room_number int,
    room_type room_types,
    price int check (price > 0),
    capacity int check (capacity > 0)
);

CREATE TABLE booking(
    id int generated always as identity primary key,
    code varchar(16) unique,
    customer_id int references customer(id),
    room_id int references room(id),
    start_date date,
    end_date date,
    constraint valid_dates check (start_date < end_date)
);

CREATE TABLE review(
    booking_id int references booking (id) primary key,
    review_comment varchar(512),
    rating int check (rating >= 1 and rating <= 5),
    review_date date
);

CREATE TABLE hotel_service(
    id int generated always as identity primary key,
    service_type hotel_services unique,
    description varchar(512),
    duration time
);

CREATE TABLE service_request(
    customer_id int references customer(id),
    service_id int references hotel_service(id),
    service_date date,
    primary key (customer_id, service_id, service_date)
);