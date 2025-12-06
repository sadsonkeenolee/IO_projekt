import CategorySwitch from "../components/CategorySwitch";
import { useState, useEffect } from "react";

export default function Favorites({ token }) {
  const [category, setCategory] = useState("film");
  const [likedItems, setLikedItems] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // placeholder: przykładowe dane
  const placeholderData = {
    film: [
      { id: 1, title: "Avatar", year: 2009 },
      { id: 2, title: "Interstellar", year: 2014 },
      { id: 3, title: "Matrix", year: 1999 },
    ],
    serial: [
      { id: 4, title: "Breaking Bad", year: 2008 },
      { id: 5, title: "Wiedźmin", year: 2019 },
    ],
    ksiazka: [
      { id: 6, title: "Władca Pierścieni", year: 1954 },
      { id: 7, title: "Harry Potter", year: 1997 },
    ],
  };

  // ładowanie danych z backendu lub fallback
  useEffect(() => {
    async function fetchLiked() {
      setLoading(true);
      setError("");
      setLikedItems([]);

      try {
        // spróbuj pobrać dane z backendu
        const res = await fetch(`http://localhost:9999/v1/users/me/liked?category=${category}`, {
          headers: { Authorization: `Bearer ${token}` },
        });

        if (!res.ok) throw new Error("Brak połączenia z serwerem");

        const data = await res.json();
        setLikedItems(data);
      } catch (err) {
        // fallback do statycznej listy
        setLikedItems(placeholderData[category]);
        setError("Nie udało się pobrać danych z serwera, użyto lokalnej listy.");
      }

      setLoading(false);
    }

    fetchLiked();
  }, [category, token]);

  // usuwa element z listy (placeholder)
  function removeFromLiked(id) {
    setLikedItems(prev => prev.filter(item => item.id !== id));

    // w przyszłości fetch DELETE do backendu:
    // await fetch(`http://localhost:9999/v1/users/me/liked/${id}`, {
    //   method: "DELETE",
    //   headers: { Authorization: `Bearer ${token}` },
    // });
  }

  return (
    <div className="max-w-4xl mx-auto mt-24 p-8 bg-slate-800 rounded-2xl shadow-xl text-white">
      <h2 className="text-2xl font-bold mb-6 text-center">Twoje polubione treści</h2>

      <CategorySwitch category={category} setCategory={setCategory} />

      {loading && <p className="text-white text-center mb-4">Wczytywanie...</p>}
      {error && <p className="text-yellow-400 text-center mb-4">{error}</p>}

      <div className="space-y-4">
        {likedItems.length === 0 && (
          <p className="text-center text-slate-400">Brak polubionych treści w tej kategorii.</p>
        )}

        {likedItems.map(item => (
          <div
            key={item.id}
            className="bg-slate-700 p-4 rounded-lg flex justify-between items-center"
          >
            <div>
              <p className="text-white font-semibold">{item.title}</p>
              {item.year && <p className="text-slate-400 text-sm">{item.year}</p>}
            </div>

            <button
              onClick={() => removeFromLiked(item.id)}
              className="px-3 py-1 rounded-md bg-red-600 hover:bg-red-700 font-medium"
            >
              Usuń z polubionych
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
