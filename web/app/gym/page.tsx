"use client";

import { useEffect, useState } from "react";
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from "recharts";

interface HistoryEntry {
  date: string;
  max_weight_kg: number;
  sets: number;
  reps: number;
}

interface PREntry {
  exercise_name: string;
  month: string;
  pr_weight_kg: number;
}

interface GymData {
  exercise: string;
  history: HistoryEntry[];
  prs: PREntry[];
}

export default function GymPage() {
  const [exercises, setExercises] = useState<string[]>([]);
  const [selected, setSelected] = useState<string>("");
  const [data, setData] = useState<GymData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch("/api/gym")
      .then((r) => r.json())
      .then((d) => {
        const list: string[] = d.exercises ?? [];
        setExercises(list);
        if (list.length > 0) setSelected(list[0]);
      })
      .catch((e: Error) => setError(e.message));
  }, []);

  useEffect(() => {
    if (!selected) return;
    fetch(`/api/gym/${encodeURIComponent(selected)}`)
      .then((r) => r.json())
      .then(setData)
      .catch((e: Error) => setError(e.message));
  }, [selected]);

  if (error) return <div className="p-8 text-red-500">Error: {error}</div>;

  return (
    <main className="p-6 max-w-3xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold">Gym</h1>

      {exercises.length === 0 ? (
        <p className="text-gray-400">Sin ejercicios registrados aun.</p>
      ) : (
        <>
          <div>
            <label className="text-sm text-gray-600 mr-2">Ejercicio:</label>
            <select
              value={selected}
              onChange={(e) => setSelected(e.target.value)}
              className="border rounded px-3 py-1 text-sm"
            >
              {exercises.map((ex) => (
                <option key={ex} value={ex}>
                  {ex}
                </option>
              ))}
            </select>
          </div>

          {data && (
            <>
              <section className="bg-white rounded-xl shadow p-4">
                <h2 className="font-semibold mb-4 capitalize">Evolucion — {selected}</h2>
                {data.history.length === 0 ? (
                  <p className="text-gray-400 text-sm">Sin datos</p>
                ) : (
                  <ResponsiveContainer width="100%" height={220}>
                    <LineChart data={data.history}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" tick={{ fontSize: 11 }} />
                      <YAxis tick={{ fontSize: 11 }} unit="kg" domain={["auto", "auto"]} />
                      <Tooltip formatter={(v: number) => `${v} kg`} />
                      <Line type="monotone" dataKey="max_weight_kg" stroke="#8b5cf6" dot={true} name="Peso max" />
                    </LineChart>
                  </ResponsiveContainer>
                )}
              </section>

              {data.prs.length > 0 && (
                <section className="bg-white rounded-xl shadow p-4">
                  <h2 className="font-semibold mb-3">PRs por mes</h2>
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="text-left text-gray-500 border-b">
                        <th className="pb-2">Mes</th>
                        <th className="pb-2">PR</th>
                      </tr>
                    </thead>
                    <tbody>
                      {data.prs.map((pr) => (
                        <tr key={pr.month} className="border-b last:border-0">
                          <td className="py-2">{pr.month}</td>
                          <td className="py-2 font-medium">{pr.pr_weight_kg} kg</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </section>
              )}
            </>
          )}
        </>
      )}
    </main>
  );
}
