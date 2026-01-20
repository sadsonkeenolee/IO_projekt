import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";

export default function Login() {
  const navigate = useNavigate();

  const [form, setForm] = useState({
    username: "",
    password: "",
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  function handleChange(e) {
    setForm({ ...form, [e.target.name]: e.target.value });
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await fetch("/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.message || "Nieprawidłowe dane logowania");
      } else {
        localStorage.setItem("token", data.access_token);
        setTimeout(() => {
            window.location.href = "/"; 
        }, 100);
      }
    } catch (err) {
      setError("Błąd połączenia z serwerem");
    }

    setLoading(false);
  }

  return (
    <div className="min-h-screen flex items-start justify-center pt-24 px-4">
      <div className="w-full max-w-md bg-slate-800 rounded-2xl shadow-xl p-8 border border-neutral-200">
        <h2 className="text-2xl font-bold text-white mb-2">Zaloguj się</h2>
        <p className="text-sm text-white mb-6">Wprowadź swoje dane, aby się zalogować.</p>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <div>
            <label className="block text-sm font-medium text-white">Login lub email</label>
            <input
              name="username"
              onChange={handleChange}
              value={form.username}
              placeholder="janek123 lub jan@przyklad.pl"
              className="mt-1 block w-full rounded-md border border-neutral-300 bg-slate-50 py-2 px-3"
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
              className="mt-1 block w-full rounded-md border border-neutral-300 bg-slate-50 py-2 px-3"
            />
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-2 px-4 bg-sky-600 hover:bg-sky-700 text-white rounded-md font-medium disabled:bg-sky-400"
          >
            {loading ? "Logowanie..." : "Zaloguj się"}
          </button>
        </form>

        <p className="mt-4 text-sm text-slate-500">
          Nie masz konta? <Link to="/register" className="text-sky-600">Zarejestruj się</Link>
        </p>
      </div>
    </div>
  );
}
