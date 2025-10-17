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

-- Popolamento tabelle
INSERT INTO customer(cf, customer_name, age, email) VALUES
('RSSMRA85M01H501Z', 'Mario', 18, 'mario.rossi@example.com'),
('VRNLCU90A12F205X', 'Lucia', 33, 'lucia.verna@example.com'),
('BNCLRD75D22L219Y', 'Ricardo', 48, 'ricardo.bianchi@example.com'),
('FRTGPP88C15D456Y', 'Giuseppe', 56, 'giuseppe.franco@example.com'),
('MNLNDR92B01F789X', 'Andrea', 31, 'andrea.milani@example.com'),
('LCPMRT77D05G123Z', 'Martina', 76, 'martina.lucchi@example.com'),
('SNCGNN81A20H321W', 'Gennaro', 72, 'gennaro.sannino@example.com'),
('PNCLRT95E11K654X', 'Roberta', 59, 'roberta.panico@example.com'),
('DVCFNC89H22L987Y', 'Francesco', 35, 'francesco.devici@example.com'),
('TMSLND84F13M432Z', 'Simone', 20, 'simone.tomasello@example.com');

INSERT INTO room(room_number, room_type, price, capacity) VALUES
(101, 'basic', 50, 2),
(102, 'suite', 120, 4),
(103, 'basic', 60, 3),
(104, 'suite', 150, 5),
(105, 'basic', 55, 2);

INSERT INTO booking(code, customer_id, room_id, start_date, end_date) VALUES
('PR001', 1, 1, '2025-01-05', '2025-01-10'),
('PR002', 2, 2, '2025-02-12', '2025-02-15'),
('PR003', 3, 3, '2025-03-01', '2025-03-03'),
('PR004', 4, 4, '2025-04-07', '2025-04-14'),
('PR005', 5, 5, '2025-05-08', '2025-05-10'),
('PR006', 6, 1, '2025-06-16', '2025-06-18'),
('PR007', 7, 2, '2025-07-19', '2025-07-22'),
('PR008', 8, 3, '2025-08-11', '2025-08-12'),
('PR009', 9, 4, '2025-09-20', '2025-09-23'),
('PR010', 10, 5, '2025-10-24', '2025-10-26');

INSERT INTO review(booking_id, review_comment, rating, review_date) VALUES
(1, 'Ottima esperienza, stanza pulita e confortevole.', 5, '2025-01-10'),
(2, 'Suite spaziosa ma un po rumorosa.', 3, '2025-02-15'),
(3, 'Camera accogliente, personale gentile.', 5, '2025-03-03'),
(4, 'Prezzo alto per la qualità offerta.', 2, '2025-04-14'),
(5, 'Molto soddisfatto del soggiorno.', 5, '2025-05-10'),
(6, 'Pulizia non perfetta, ma buona posizione.', 3, '2025-06-18'),
(7, 'Servizio eccellente, ci tornerò sicuramente.', 5, '2025-07-22');

INSERT INTO hotel_service(service_type, description, duration) VALUES
('cleaning', 'Servizio di pulizia della stanza.', '01:00:00'),
('room_service', 'Servizio di ordinazione cibo in camera.', '00:30:00'),
('massage', 'Massaggio rilassante in camera.', '01:30:00');

INSERT INTO service_request(customer_id, service_id, service_date) VALUES
(1, 1, '2025-10-06'),
(2, 3, '2025-01-11'),
(3, 2, '2025-02-12'),
(4, 3, '2025-03-02'),
(5, 1, '2025-04-12'),
(6, 2, '2025-05-08'),
(7, 3, '2025-06-16'),
(8, 1, '2025-07-20');