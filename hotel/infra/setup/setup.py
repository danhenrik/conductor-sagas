import psycopg2
import time

class pg_conn:
  def __init__(self):
    try:
      # conn = psycopg2.connect("host=database-hotel port=5432 user=saga-hotel password=saga-hotel dbname=saga-hotel")
      conn = psycopg2.connect("host=localhost port=5433 user=saga-hotel password=saga-hotel dbname=saga-hotel")
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

pg_conn.SQLCmd("""CREATE TABLE hotels (
    hotel_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    rating INT,
    rooms_available INT NOT NULL
  );""")

pg_conn.SQLCmd("""CREATE TABLE hotel_bookings (
    booking_id VARCHAR(50) PRIMARY KEY,
    hotel_id VARCHAR(50) NOT NULL,
    check_in_date TIMESTAMP NOT NULL,
    check_out_date TIMESTAMP NOT NULL,
    customer_name VARCHAR(100) NOT NULL,
    customer_email VARCHAR(100) NOT NULL,
    room_number INT NOT NULL,
    booking_status VARCHAR(50) NOT NULL,
    booking_time TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (hotel_id) REFERENCES hotels(hotel_id) ON DELETE CASCADE
  );""")

pg_conn.SQLCmd("""CREATE UNIQUE INDEX unique_active_booking 
  ON hotel_bookings (hotel_id, room_number) 
  WHERE booking_status = 'active';""")