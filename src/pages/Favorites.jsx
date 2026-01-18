import CategorySwitch from "../components/CategorySwitch";
import { useState, useEffect } from "react";

export default function Favorites({ token }) {
  const [category, setCategory] = useState("film");
  const [allLiked, setAllLiked] = useState([]); 
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    async function fetchAllLiked() {
      if (!token) {
        setError("Musisz być zalogowany, aby zobaczyć ulubione.");
        return;
      }

      setLoading(true);
      try {
        const res = await fetch(`http://localhost:9999/v1/users/me/liked`, {
          headers: { Authorization: `Bearer ${token}` },
        });

        if (!res.ok) throw new Error("Nie udało się pobrać danych.");

        const data = await res.json(); 
        setAllLiked(data); 
      } catch (err) {
        setError("Błąd ładowania ulubionych.");
      } finally {
        setLoading(false);
      }
    }

    fetchAllLiked();
  }, [token]);

  const filteredItems = allLiked.filter(item => item.type === category);

  async function handleUnlike(id) {
    try {
      const response = await fetch("http://localhost:9999/v1/api/likes", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          token: token,
          type: category,
          id: id,
          event: 'dislike' 
        }),
      });

      if (response.ok) {
        setAllLiked(prev => prev.filter(item => item.id !== id));
      }
    } catch (err) {
      console.error("Błąd podczas usuwania:", err);
    }
  }

  return (
    <div className="max-w-4xl mx-auto mt-24 p-8 bg-slate-800 rounded-2xl shadow-xl text-white">
      <h2 className="text-2xl font-bold mb-6 text-center">Twoje polubione treści</h2>

      <CategorySwitch category={category} setCategory={setCategory} />

      {loading && <p className="text-center my-4">Wczytywanie Twojej listy...</p>}
      {error && <p className="text-red-400 text-center my-4">{error}</p>}

      <div className="space-y-4 mt-6">
        {filteredItems.length === 0 && !loading && (
          <p className="text-center text-slate-400">
            Brak polubionych {category === 'ksiazka' ? 'książek' : category + "ów"}.
          </p>
        )}

        {filteredItems.map(item => (
          <div
            key={item.id}
            className="bg-slate-700 p-4 rounded-lg flex justify-between items-center animate-fadeIn"
          >
            <div>
              <p className="text-white font-semibold">{item.title}</p>
              <p className="text-slate-400 text-sm">{item.year}</p>
            </div>

            <button
              onClick={() => handleUnlike(item.id)}
              className="px-4 py-2 rounded-md bg-red-500/20 hover:bg-red-600 text-red-400 hover:text-white border border-red-500/50 transition-all text-sm font-medium"
            >
              Odlub 💔
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}