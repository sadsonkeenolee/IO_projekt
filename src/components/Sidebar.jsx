import { Link } from "react-router-dom";

export default function Navbar({ isLoggedIn }) {
  return (
    <aside className="w-full bg-slate-800 p-4 shadow-xl border-b border-slate-700 flex items-center justify-between">
      
      <h2 className="text-xl font-bold mr-8">📚 Menu</h2>

      <nav className="flex space-x-6 items-center">
        <Link to="/" className="hover:text-blue-400">🏠 Strona główna</Link>
        <Link to="/register" className="hover:text-blue-400">🔐 Zarejestruj</Link>
        
        {isLoggedIn ? (
          <Link to="/account" className="hover:text-blue-400">👤 Szczegóły konta</Link>
        ) : (
          <Link to="/login" className="hover:text-blue-400">🔐 Zaloguj</Link>
        )}
        
        <Link to="/favorites" className="hover:text-blue-400">🔐 Polubione</Link>
        <Link to="/about" className="hover:text-blue-400">🧠 O algorytmie</Link>
        <Link to="/sources" className="hover:text-blue-400">📊 Źródła</Link>
      </nav>
    </aside>
  );
}