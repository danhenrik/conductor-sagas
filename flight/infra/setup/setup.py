import psycopg2
import time

class pg_conn:
  def __init__(self):
    try:
      # conn = psycopg2.connect("host=database-flight user=saga-flight password=saga-flight dbname=saga-flight")
      conn = psycopg2.connect("host=localhost user=saga-flight password=saga-flight dbname=saga-flight")
      self.conn = conn
      self.c = conn.cursor()
      print("Successfully connected")
    except Exception as e: 
      print(e.__str__())

  def SQLCmd(self, cmd):
    try:
      print("Executing : \"" + cmd + '"')
      self.c.execute(cmd)
      self.conn.commit()
    except Exception as e:
      print("Execution failed!\nError: "+ e.__str__())
      self.conn.commit()

print("Waiting 10 seconds for services to start")
time.sleep(10)

# Create Postgres Tables
pg_conn = pg_conn()

pg_conn.SQLCmd("""CREATE TABLE flights (
    flight_id VARCHAR(50) PRIMARY KEY,
    airline VARCHAR(100) NOT NULL,
    origin VARCHAR(100) NOT NULL,
    destination VARCHAR(100) NOT NULL,
    departure_time TIMESTAMP NOT NULL,
    arrival_time TIMESTAMP NOT NULL,
    capacity INT NOT NULL,
    available_seats INT NOT NULL
  );""")

pg_conn.SQLCmd("""CREATE TABLE flight_bookings (
    booking_id VARCHAR(50) PRIMARY KEY,
    flight_id VARCHAR(50) NOT NULL,
    customer_name VARCHAR(100) NOT NULL,
    customer_email VARCHAR(100) NOT NULL,
    seat_number INT NOT NULL,
    booking_status VARCHAR(50) NOT NULL,
    booking_time TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (flight_id) REFERENCES flights(flight_id) ON DELETE CASCADE
  );""")

pg_conn.SQLCmd("""CREATE UNIQUE INDEX unique_active_flight_seat 
  ON flight_bookings (flight_id, seat_number) 
  WHERE booking_status = 'active';""")


