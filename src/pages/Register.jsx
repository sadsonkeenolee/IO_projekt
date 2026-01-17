import { useState } from "react";
import { Link } from "react-router-dom";

export default function Register() {
  const [form, setForm] = useState({
    username: "",
    email: "",
    password: "",
    birthday: "",
    gender: "",
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  function handleChange(e) {
    setForm({ ...form, [e.target.name]: e.target.value });
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setLoading(true);
    setError("");
    setSuccess("");

    try {
      const res = await fetch("/api/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.message || "Wystąpił błąd");
      } else {
        setSuccess("Konto zostało utworzone!");
        localStorage.setItem("token", data.token);
        navigate("/");
      }
    } catch (err) {
      setError("Błąd połączenia z serwerem");
    }

    setLoading(false);
  }

  return (
    <div className="min-h-screen flex items-start justify-center pt-24 px-4">
      <div className="w-full max-w-md bg-slate-800 rounded-2xl shadow-xl p-8 border border-slate-200">
        <h2 className="text-2xl font-bold text-white mb-2">Załóż konto</h2>
        <p className="text-sm text-white mb-6">Szybka rejestracja.</p>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <div>
            <label className="block text-sm font-medium text-white">Nazwa użytkownika</label>
            <input
              name="username"
              onChange={handleChange}
              value={form.username}
              placeholder="Janek123"
              className="mt-1 block w-full rounded-md border border-slate-300 bg-slate-50 py-2 px-3"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-white">Email</label>
            <input
              name="email"
              type="email"
              onChange={handleChange}
              value={form.email}
              placeholder="jan@przyklad.pl"
              className="mt-1 block w-full rounded-md border border-slate-300 bg-slate-50 py-2 px-3"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-white">Hasło</label>
            <input
              name="password"
              type="password"
              onChange={handleChange}
              value={form.password}
              placeholder="••••••••"
              className="mt-1 block w-full rounded-md border border-slate-300 bg-slate-50 py-2 px-3"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-white">Data urodzenia</label>
            <input
              name="birthday"
              type="date"
              onChange={handleChange}
              value={form.birthday}
              className="mt-1 block w-full rounded-md border border-slate-300 bg-slate-50 py-2 px-3"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-white">Płeć</label>
            <select
              name="gender"
              onChange={handleChange}
              value={form.gender}
              className="mt-1 block w-full rounded-md border border-slate-300 bg-slate-50 py-2 px-3"
            >
              <option value="">Wybierz...</option>
              <option value="male">Mężczyzna</option>
              <option value="female">Kobieta</option>
            </select>
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}
          {success && <p className="text-green-400 text-sm">{success}</p>}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-2 px-4 bg-sky-600 hover:bg-sky-700 text-white rounded-md font-medium disabled:bg-sky-400"
          >
            {loading ? "Rejestrowanie..." : "Zarejestruj się"}
          </button>
        </form>

        <p className="mt-4 text-sm text-slate-500">
          Masz już konto? <Link to="/login" className="text-sky-600">Zaloguj się</Link>
        </p>
      </div>
    </div>
  );
}