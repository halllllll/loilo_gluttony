DROP TABLE IF EXISTS info;
CREATE TABLE info(
  created_at TIMESTAMP DEFAULT (datetime(CURRENT_TIMESTAMP,'localtime')),
  goon BOOLEAN
);

DROP TABLE IF EXISTS students;
CREATE TABLE students(
  school_name TEXT NOT NULL,
  loilo_user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  kana TEXT,
  password TEXT,
  google_account_id TEXT,
  ms_account_id TEXT, 
  grade TEXT,
  class TEXT
);

DROP TABLE IF EXISTS teachers;
CREATE TABLE teachers(
  school_name TEXT NOT NULL,
  loilo_user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  kana TEXT,
  password TEXT,
  google_account_id TEXT,
  ms_account_id TEXT
);