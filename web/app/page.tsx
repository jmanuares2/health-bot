"use client";

import { useEffect, useState } from "react";

interface FoodEntry {
  id: number;
  meal_type: string;
  description: string;
  calories: number;
  protein_g: number;
}

interface StrengthSession {
  id: number;
  exercise_name: string;
  sets: number;
  reps: number;
  weight_kg: number;
}

interface CardioSession {
  id: number;
  activity: string;
  duration_min: number;
  distance_km: number;
  calories_burned: number;
}

interface TodayData {
  date: string;
  calorie_goal: number;
  calories: number;
  protein_g: number;
  carbs_g: number;
  fat_g: number;
  fiber_g: number;
  water_liters: number;
  food_entries: FoodEntry[];
  strength: StrengthSession[];
  cardio: CardioSession[];
}

export default function TodayPage() {
  const [data, setData] = useState<TodayData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch("/api/today")
      .then((r) => r.json())
      .then(setData)
      .catch((e: Error) => setError(e.message));
  }, []);

  if (error) return <div className="p-8 text-red-500">Error: {error}</div>;
  if (!data) return <div className="p-8 text-gray-400">Cargando...</div>;

  const pct = Math.min(100, Math.round((data.calories / data.calorie_goal) * 100));
  const remaining = data.calorie_goal - data.calories;

  return (
    <main className="p-6 max-w-2xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold">Hoy — {data.date}</h1>

      <section className="bg-white rounded-xl shadow p-4 space-y-2">
        <div className="flex justify-between text-sm text-gray-600">
          <span>{data.calories} kcal consumidas</span>
          <span>{data.calorie_goal} kcal objetivo</span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-4">
          <div
            className={`h-4 rounded-full transition-all ${pct >= 100 ? "bg-red-500" : "bg-green-500"}`}
            style={{ width: `${pct}%` }}
          />
        </div>
        <p className="text-sm text-gray-500">
          {remaining >= 0
            ? `Quedan ${remaining} kcal`
            : `Excedido en ${Math.abs(remaining)} kcal`}
        </p>
      </section>

      <section className="bg-white rounded-xl shadow p-4">
        <h2 className="font-semibold mb-3">Macros</h2>
        <div className="grid grid-cols-4 gap-3 text-center">
          {[
            { label: "Proteinas", value: data.protein_g, color: "text-blue-600" },
            { label: "Carbos", value: data.carbs_g, color: "text-yellow-600" },
            { label: "Grasas", value: data.fat_g, color: "text-orange-600" },
            { label: "Fibra", value: data.fiber_g, color: "text-green-600" },
          ].map((m) => (
            <div key={m.label} className="bg-gray-50 rounded-lg p-2">
              <p className={`text-xl font-bold ${m.color}`}>{m.value.toFixed(0)}g</p>
              <p className="text-xs text-gray-500">{m.label}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="bg-white rounded-xl shadow p-4 flex items-center gap-4">
        <span className="text-3xl">💧</span>
        <div>
          <p className="text-2xl font-bold text-blue-500">{data.water_liters.toFixed(1)} L</p>
          <p className="text-sm text-gray-500">Agua consumida hoy</p>
        </div>
      </section>

      {data.food_entries.length > 0 && (
        <section className="bg-white rounded-xl shadow p-4">
          <h2 className="font-semibold mb-3">Comidas</h2>
          <ul className="divide-y">
            {data.food_entries.map((f) => (
              <li key={f.id} className="py-2 flex justify-between text-sm">
                <span>
                  <span className="text-gray-400 mr-2 capitalize">{f.meal_type}</span>
                  {f.description}
                </span>
                <span className="font-medium">{f.calories} kcal</span>
              </li>
            ))}
          </ul>
        </section>
      )}

      {(data.strength.length > 0 || data.cardio.length > 0) && (
        <section className="bg-white rounded-xl shadow p-4">
          <h2 className="font-semibold mb-3">Entrenamientos</h2>
          <ul className="divide-y text-sm">
            {data.strength.map((s) => (
              <li key={s.id} className="py-2">
                💪 {s.exercise_name} {s.sets}×{s.reps} @ {s.weight_kg}kg
              </li>
            ))}
            {data.cardio.map((c) => (
              <li key={c.id} className="py-2">
                🏃 {c.activity} — {c.distance_km}km / {c.duration_min}min · {c.calories_burned} kcal
              </li>
            ))}
          </ul>
        </section>
      )}
    </main>
  );
}
