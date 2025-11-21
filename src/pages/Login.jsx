import { Link } from "react-router-dom";

export default function Login() {
  return (
    <div className="min-h-screen flex items-start justify-center pt-24 px-4">
      <div className="w-full max-w-md bg-slate-800 rounded-2xl shadow-xl p-8 border border-slate-200">
        <h2 className="text-2xl font-bold text-white mb-2">Zaloguj się</h2>
        <p className="text-sm text-white mb-6">Szybkie logowanie — wygląd bez funkcji.</p>

        <form className="space-y-4" onSubmit={(e)=>e.preventDefault()}>
          <div>
            <label className="block text-sm font-medium text-white">Email</label>
            <input readOnly placeholder="jan@przyklad.pl" className="mt-1 block w-full rounded-md border border-slate-300 bg-slate-50 py-2 px-3" />
          </div>

          <div>
            <label className="block text-sm font-medium text-white">Hasło</label>
            <input readOnly placeholder="••••••••" className="mt-1 block w-full rounded-md border border-slate-300 bg-slate-50 py-2 px-3" />
          </div>

          <button type="button" className="w-full py-2 px-4 bg-sky-600 hover:bg-sky-700 text-white rounded-md font-medium">
            Zaloguj się
          </button>
        </form>

        <p className="mt-4 text-sm text-slate-500">
          Nie masz konta? <Link to="/register" className="text-sky-600">Zarejestruj się</Link>
        </p>
      </div>
    </div>
  );
}