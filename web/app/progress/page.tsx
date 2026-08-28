"use client";

import { useEffect, useState } from "react";
import {
  LineChart, Line, BarChart, Bar, XAxis, YAxis,
  CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine,
} from "recharts";

interface DayCalories {
  date: string;
  total_calories: number;
}

interface WeightEntry {
  date: string;
  weight_kg: number;
}

interface ProgressData {
  month: string;
  calories: DayCalories[];
  body_weight: WeightEntry[];
  adherence: number;
}

export default function ProgressPage() {
  const [data, setData] = useState<ProgressData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [month, setMonth] = useState(() => new Date().toISOString().slice(0, 7));

  useEffect(() => {
    fetch(`/api/progress?month=${month}`)
      .then((r) => r.json())
      .then(setData)
      .catch((e: Error) => setError(e.message));
  }, [month]);

  if (error) return <div className="p-8 text-red-500">Error: {error}</div>;

  return (
    <main className="p-6 max-w-3xl mx-auto space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Progreso</h1>
        <input
          type="month"
          value={month}
          onChange={(e) => setMonth(e.target.value)}
          className="border rounded px-3 py-1 text-sm"
        />
      </div>

      {data && (
        <>
          <section className="bg-white rounded-xl shadow p-4">
            <h2 className="font-semibold mb-4">Peso corporal</h2>
            {data.body_weight.length === 0 ? (
              <p className="text-gray-400 text-sm">Sin datos</p>
            ) : (
              <ResponsiveContainer width="100%" height={200}>
                <LineChart data={data.body_weight}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" tick={{ fontSize: 11 }} />
                  <YAxis domain={["auto", "auto"]} tick={{ fontSize: 11 }} unit="kg" />
                  <Tooltip formatter={(v: number) => `${v} kg`} />
                  <Line type="monotone" dataKey="weight_kg" stroke="#3b82f6" dot={true} />
                </LineChart>
              </ResponsiveContainer>
            )}
          </section>

          <section className="bg-white rounded-xl shadow p-4">
            <h2 className="font-semibold mb-4">Calorias diarias</h2>
            {data.calories.length === 0 ? (
              <p className="text-gray-400 text-sm">Sin datos</p>
            ) : (
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={data.calories}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip />
                  <ReferenceLine y={2000} stroke="#ef4444" strokeDasharray="4 4" label={{ value: "Objetivo", fontSize: 11 }} />
                  <Bar dataKey="total_calories" fill="#22c55e" name="Calorias" />
                </BarChart>
              </ResponsiveContainer>
            )}
          </section>

          <section className="bg-white rounded-xl shadow p-4 flex items-center gap-4">
            <div>
              <p className="text-4xl font-bold text-green-600">{data.adherence.toFixed(0)}%</p>
              <p className="text-sm text-gray-500">Adherencia mensual (dias en objetivo)</p>
            </div>
          </section>
        </>
      )}

      {!data && <p className="text-gray-400">Cargando...</p>}
    </main>
  );
}
