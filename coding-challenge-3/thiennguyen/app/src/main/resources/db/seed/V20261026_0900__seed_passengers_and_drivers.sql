-- Seed data for local runs and scripts/demo.sh. Automated tests create their own records in isolated areas.
-- Distances are from the AC-01 example pickup (10.7769, 106.7009), Ho Chi Minh City.

INSERT INTO passengers (id, name) VALUES
    ('P001', 'Passenger One'),
    ('P002', 'Passenger Two'),
    ('P003', 'Passenger Three'),
    ('P004', 'Passenger Four'),
    ('P005', 'Passenger Five');

INSERT INTO drivers (id, availability, latitude, longitude, location_updated_at, updated_at) VALUES
    ('D001', 'AVAILABLE', 10.784095, 106.7009, now(), now()), -- 0.8 km (AC-04 "Driver A")
    ('D002', 'AVAILABLE', 10.790390, 106.7009, now(), now()), -- 1.5 km (AC-04 "Driver B")
    ('D003', 'AVAILABLE', 10.805678, 106.7009, now(), now()), -- 3.2 km (AC-04 "Driver C", outside 2 km)
    ('D004', 'OFFLINE',   10.779598, 106.7009, now(), now()), -- 0.3 km but offline
    ('D005', 'AVAILABLE', 10.787692, 106.7009, now(), now()), -- 1.2 km
    ('D006', 'AVAILABLE', 10.793088, 106.7009, now(), now()), -- 1.8 km
    ('D007', 'AVAILABLE', 10.799383, 106.7009, now(), now()), -- 2.5 km, outside 2 km
    ('D008', 'OFFLINE',   NULL,      NULL,     NULL,  now()),
    ('D009', 'OFFLINE',   NULL,      NULL,     NULL,  now()),
    ('D010', 'OFFLINE',   NULL,      NULL,     NULL,  now());

-- Dispatch's supply copy of the same drivers, as if their driver-events had already been consumed (ADR-002).
INSERT INTO driver_supply (driver_id, online, latitude, longitude, updated_at)
SELECT id, availability = 'AVAILABLE', latitude, longitude, now() FROM drivers;
