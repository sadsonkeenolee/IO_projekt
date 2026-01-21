import CategorySwitch from "../components/CategorySwitch";
import { useState, useEffect } from "react";

export default function Favorites({ token }) {
  const [category, setCategory] = useState("film");
  const [allLiked, setAllLiked] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const typeMap = {
    film: "tv",
    ksiazka: "book",
  };

  useEffect(() => {
    const controller = new AbortController();
    
    async function fetchData() {
      if (!token) return;

      // KLUCZOWA ZMIANA: Czyścimy listę i błędy NATYCHMIAST po zmianie kategorii
      setAllLiked([]);
      setError("");
      setLoading(true);

      try {
        // 1. Pobranie listy ID polubionych rzeczy
        const resp = await fetch("/v1/auth/event/pull", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            access_token: token,
            event: "like",
            type: typeMap[category],
          }),
          signal: controller.signal
        });

        if (!(resp?.ok || resp?.status === 302)) {
          throw new Error("Nie udało się pobrać listy polubionych.");
        }

        const data = await resp.json();
        const items = data.content?.items || [];

        if (items.length === 0) {
          setAllLiked([]);
          setLoading(false);
          return;
        }

        // 2. Pobranie szczegółów dla każdego elementu
        const results = [];
        const endpoint = category === "film" ? "tv" : "book";

        for (const item of items) {
          try {
            const details = await fetch(`/v1/api/${endpoint}/id/${item.id}`, {
              signal: controller.signal
            });
            
            if (!(details?.ok || details?.status === 302)) continue;

            const content = await details.json();
            if (content) {
              results.push(content);
            }
          } catch (err) {
            if (err.name !== 'AbortError') console.error("Błąd detali:", err);
          }
        }

        setAllLiked(results);
      } catch (err) {
        if (err.name !== 'AbortError') {
          console.error("Błąd główny:", err);
          setError("Wystąpił błąd podczas ładowania danych.");
        }
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    }

    fetchData();

    // Cleanup przy zmianie kategorii lub odmontowaniu komponentu
    return () => controller.abort();
  }, [token, category]);

  async function handleUnlike(id) {
    try {
      const response = await fetch("/v1/auth/event/push", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          access_token: token,
          event: "dislike",
          type: typeMap[category],
          id: id.toString(),
        }),
      });

      if (response.ok || response.status === 302) {
        // Filtrowanie uwzględniające oba typy ID (movie_id dla tv, id dla book)
        setAllLiked((prev) => prev.filter((item) => {
          const itemId = item.content.movie_id || item.content.id;
          return itemId !== id;
        }));
      }
    } catch (err) {
      console.error("Błąd usuwania:", err);
    }
  }

  // Pomocnik do ujednolicenia wyświetlania danych
  const getDisplayData = (item) => ({
    id: item.content.movie_id || item.content.id,
    title: item.content.title || item.content.name || "Brak tytułu",
    date: item.content.release_date || item.content.published_date || "Brak daty"
  });

  return (
    <div className="max-w-4xl mx-auto mt-24 p-8 bg-slate-800 rounded-2xl shadow-xl text-white">
      <h2 className="text-2xl font-bold mb-6 text-center">Twoje polubione treści</h2>

      <CategorySwitch category={category} setCategory={setCategory} />

      {loading && (
        <div className="flex justify-center items-center my-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mr-3"></div>
          <p className="text-blue-400">Pobieranie: {category === "film" ? "Filmy" : "Książki"}...</p>
        </div>
      )}
      
      {error && <p className="text-red-400 text-center my-4 p-2 bg-red-900/20 rounded">{error}</p>}

      <div className="space-y-4 mt-6">
        {!loading && allLiked.length === 0 && !error && (
          <p className="text-center text-slate-400 py-10 border-2 border-dashed border-slate-700 rounded-xl">
            Brak pozycji w kategorii {category === "film" ? "filmy" : "książki"}.
          </p>
        )}

        {allLiked.map((item) => {
          const { id, title, date } = getDisplayData(item);
          return (
            <div 
              key={`${category}-${id}`} 
              className="bg-slate-700 p-4 rounded-lg flex justify-between items-center hover:bg-slate-600 transition-all transform hover:scale-[1.01] shadow-sm"
            >
              <div>
                <p className="text-white font-semibold">{title}</p>
                <p className="text-slate-400 text-sm">{date}</p>
              </div>
              <button
                onClick={() => handleUnlike(id)}
                className="px-4 py-2 rounded-md bg-red-500/10 hover:bg-red-600 text-red-500 hover:text-white border border-red-500/50 transition-all text-sm font-medium"
              >
                Odlub 💔
              </button>
            </div>
          );
        })}
      </div>
    </div>
  );
}