import { Link } from "react-router-dom";
import Logout from "./Logout";

export default function Navbar({ isLoggedIn }) {
  // Wspólny styl dla standardowych linków (żeby nie powtarzać kodu)
  const linkStyle = "flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800 transition-all";

  return (
    <nav className="flex items-center gap-2">
      
      {/* 1. Strona główna zawsze widoczna */}
      <Link to="/" className={linkStyle}>
        🏠 Strona główna
      </Link>

      {/* Pionowa kreska oddzielająca Home od reszty */}
      <div className="h-5 w-px bg-slate-700 mx-1"></div>

      {isLoggedIn ? (
        // --- WERSJA DLA ZALOGOWANEGO ---
        <>
          <Link to="/favorites" className={linkStyle}>
            ❤️ Polubione
          </Link>
          <Link to="/account" className={linkStyle}>
            👤 Konto
          </Link>
          <Link to="/about" className={linkStyle}>
            🧠 Algorytm
          </Link>
          <Link to="/sources" className={linkStyle}>
            📊 Źródła
          </Link>
          
          {/* Logout jako przycisk, oddzielony marginesem */}
          <div className="ml-2 pl-2 border-l border-neutral-700">
             <Logout />
          </div>
        </>
      ) : (
        // --- WERSJA DLA GOŚCIA ---
        <>
          {/* Linki informacyjne */}
          <Link to="/about" className={linkStyle}>
            🧠 Algorytm
          </Link>
          <Link to="/sources" className={linkStyle}>
            📊 Źródła
          </Link>

          {/* Sekcja logowania przesunięta nieco w prawo */}
          <div className="flex items-center gap-2 ml-4 pl-4 border-l border-neutral-700">
            <Link to="/login" className={linkStyle}>
              🔐 Zaloguj
            </Link>
            
            {/* Wyróżniony przycisk rejestracji */}
            <Link 
              to="/register" 
              className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-bold text-white bg-emerald-600 hover:bg-emerald-500 hover:scale-105 transition-all shadow-lg shadow-emerald-900/20"
            >
              📝 Zarejestruj
            </Link>
          </div>
        </>
      )}
    </nav>
  );
}
