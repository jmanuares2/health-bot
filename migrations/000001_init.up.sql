-- Configuración del usuario
CREATE TABLE users (
  id              SERIAL PRIMARY KEY,
  phone           VARCHAR(20) UNIQUE NOT NULL,
  calorie_goal    INT NOT NULL DEFAULT 2000,
  protein_goal_g  INT,
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Comidas
CREATE TABLE food_entries (
  id          SERIAL PRIMARY KEY,
  user_id     INT REFERENCES users(id),
  date        DATE NOT NULL,
  meal_type   VARCHAR(20),
  description TEXT NOT NULL,
  calories    INT NOT NULL,
  protein_g   DECIMAL(6,1),
  carbs_g     DECIMAL(6,1),
  fat_g       DECIMAL(6,1),
  fiber_g     DECIMAL(6,1),
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Ejercicios de fuerza
CREATE TABLE strength_sessions (
  id              SERIAL PRIMARY KEY,
  user_id         INT REFERENCES users(id),
  date            DATE NOT NULL,
  exercise_name   VARCHAR(100) NOT NULL,
  sets            INT,
  reps            INT,
  weight_kg       DECIMAL(6,2),
  calories_burned INT,
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Cardio
CREATE TABLE cardio_sessions (
  id              SERIAL PRIMARY KEY,
  user_id         INT REFERENCES users(id),
  date            DATE NOT NULL,
  activity        VARCHAR(100) NOT NULL,
  duration_min    INT,
  distance_km     DECIMAL(6,2),
  calories_burned INT,
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Peso corporal
CREATE TABLE body_weight (
  id         SERIAL PRIMARY KEY,
  user_id    INT REFERENCES users(id),
  date       DATE NOT NULL,
  weight_kg  DECIMAL(5,2) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Hidratación
CREATE TABLE water_logs (
  id         SERIAL PRIMARY KEY,
  user_id    INT REFERENCES users(id),
  date       DATE NOT NULL,
  liters     DECIMAL(4,2) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
