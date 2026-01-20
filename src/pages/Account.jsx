import { useState, useEffect } from "react";

export default function Account({ token }) {
  const [user, setUser] = useState({
    username: "jan",
    email: "jan@przyklad.pl",
    gender: "mężczyzna",
    birthday: "1990-01-01",
  });

  const [form, setForm] = useState({ ...user, password: "" });
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");

  function handleChange(e) {
    const { name, value } = e.target;
    setForm(prev => ({ ...prev, [name]: value }));
  }

  async function handleSave() {
    setLoading(true);
    setMessage("");

    try {
      // tutaj normalnie fetch do backendu np.:
      // await fetch("http://localhost:9999/v1/auth/update", {
      //   method: "PUT",
      //   headers: { 
      //       "Content-Type": "application/json",
      //       "Authorization": `Bearer ${token}`
      //   },
      //   body: JSON.stringify(form)
      // });

      // placeholder: zapisujemy lokalnie
      setUser({ ...form });
      setForm(prev => ({ ...prev, password: "" }));
      setMessage("Zapisano zmiany!");
    } catch (err) {
      setMessage("Nie udało się zapisać zmian.");
    }

    setLoading(false);
  }

  // symulacja usunięcia konta
  async function handleDelete() {
    if (!confirm("Na pewno chcesz usunąć konto?")) return;

    setLoading(true);
    setMessage("");

    try {
      setUser(null);
      setMessage("Konto zostało usunięte (placeholder).");
    } catch (err) {
      setMessage("Nie udało się usunąć konta.");
    }

    setLoading(false);
  }

  if (!user) {
    return <p className="text-center text-white mt-10">Brak danych użytkownika.</p>;
  }

  return (
    <div className="max-w-xl mx-auto p-8 bg-slate-800 rounded-2xl shadow-xl text-white mt-24">
      <h2 className="text-2xl font-bold mb-6">Szczegóły konta</h2>

      <div className="space-y-4">
        <div>
          <label className="block mb-1">Nazwa użytkownika</label>
          <input
            type="text"
            name="username"
            value={form.username}
            onChange={handleChange}
            className="w-full px-3 py-2 rounded-md bg-slate-700 border border-neutral-600"
          />
        </div>

        <div>
          <label className="block mb-1">E-mail</label>
          <input
            type="email"
            name="email"
            value={form.email}
            readOnly
            className="w-full px-3 py-2 rounded-md bg-slate-600 border border-neutral-500 cursor-not-allowed"
          />
        </div>

        <div>
          <label className="block mb-1">Hasło</label>
          <input
            type="password"
            name="password"
            value={form.password}
            onChange={handleChange}
            placeholder="••••••••"
            className="w-full px-3 py-2 rounded-md bg-slate-700 border border-neutral-600"
          />
        </div>

        <div>
          <label className="block mb-1">Płeć</label>
          <select
            name="gender"
            value={form.gender}
            onChange={handleChange}
            className="w-full px-3 py-2 rounded-md bg-slate-700 border border-neutral-600"
          >
            <option value="mężczyzna">Mężczyzna</option>
            <option value="kobieta">Kobieta</option>
            <option value="inne">Inne</option>
          </select>
        </div>

        <div>
          <label className="block mb-1">Data urodzenia</label>
          <input
            type="date"
            name="birthday"
            value={form.birthday}
            onChange={handleChange}
            className="w-full px-3 py-2 rounded-md bg-slate-700 border border-neutral-600"
          />
        </div>

        <div className="flex gap-4 mt-6">
          <button
            onClick={handleSave}
            disabled={loading}
            className="px-4 py-2 bg-sky-600 hover:bg-sky-700 rounded-md font-medium"
          >
            Zapisz zmiany
          </button>

          <button
            onClick={handleDelete}
            disabled={loading}
            className="px-4 py-2 bg-red-600 hover:bg-red-700 rounded-md font-medium"
          >
            Usuń konto
          </button>
        </div>

        {message && <p className="mt-4 text-yellow-400">{message}</p>}
      </div>
    </div>
  );
}
