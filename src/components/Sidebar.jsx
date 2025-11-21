import { Link } from "react-router-dom";

export default function Sidebar() {
  return (
    <aside className="w-64 bg-slate-800 p-6 shadow-xl border-r border-slate-700">
      <h2 className="text-xl font-bold mb-8">📚 Menu</h2>

      <nav className="space-y-4">

        <Link to="/" className="block hover:text-blue-400">🏠 Strona główna</Link>
        <Link to="/register" className="block hover:text-blue-400">🔐 Zarejestruj</Link>
        <Link to="/login" className="block hover:text-blue-400">🔐 Zaloguj</Link>
        <Link to="/favorites" className="block hover:text-blue-400">🔐 Polubione treści</Link>
        <Link to="/account" className="block hover:text-blue-400">👤 Szczegóły konta</Link>
        <Link to="/about" className="block hover:text-blue-400">🧠 O algorytmie</Link>
        <Link to="/sources" className="block hover:text-blue-400">📊 Źródła danych</Link>

      </nav>
    </aside>
  );
}