Create Table If Not Exists "bookings" (
  "id" UUID Not Null,
  "booking_code" Character Varying (50) Not Null,
  "user_id" UUID Not Null,
  "total_amount" Decimal(15, 2) Not Null Default 0,
  "status" Character Varying (20) Not Null Default 'PENDING', -- PENDING, CONFIRMED, CANCELLED, COMPLETED
  "payment_status" Character Varying (20) Not Null Default 'UNPAID',
  "created_at" BigInt Not Null Default 0,
  "updated_at" BigInt Null,
  "deleted_at" BigInt Null,

  Constraint "pk_bookings" Primary Key ("id"),
  Constraint "unq_bookings_booking_code" Unique ("booking_code")
);

Comment On Column "bookings"."status" Is '- PENDING
- CONFIRMED
- CANCELLED
- COMPLETED';

Create Table If Not Exists "booking_details" (
  "id" UUID Not Null,
  "booking_id" UUID Not Null,
  "product_id" UUID NOT NULL,
  "product_name" Character Varying (100),
  "qty" Int Not Null Default 1,
  "price_per_unit" Decimal(15, 2) Not Null,
  "sub_total" Decimal(15, 2) Not Null,
  "created_at" BigInt Not Null Default 0,
  "updated_at" BigInt Null,

  Constraint "pk_booking_details" Primary Key ("id"),
  Constraint "fk_booking_details_bookings" Foreign Key ("booking_id") References "bookings" ("id") On Delete Cascade
);