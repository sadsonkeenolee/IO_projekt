import CategorySwitch from "../components/CategorySwitch";
import { useState, useEffect } from "react";

export default function Favorites({ token }) {
  const [category, setCategory] = useState("filmy i seriale");
  const [allLiked, setAllLiked] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const typeMap = {
    "filmy i seriale": "tv",
    ksiazki: "book",
  };

  useEffect(() => {
    const controller = new AbortController();

    async function fetchData() {
      if (!token) return;

      setLoading(true);
      setError("");
      setAllLiked([]);

      try {
        const currentType = typeMap[category];

        // 1. Lista polubień
        const listResp = await fetch("/v1/auth/event/pull", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            access_token: token,
            event: "like",
            type: currentType,
          }),
          signal: controller.signal,
        });

        if (!listResp.ok && listResp.status !== 302) throw new Error("Błąd listy");

        const listData = await listResp.json();
        const rawItems = listData.content?.items || [];
        const filteredItems = rawItems.filter((item) => item.type === currentType);

        if (filteredItems.length === 0) {
          setAllLiked([]);
          setLoading(false);
          return;
        }

        // 2. Pobieranie szczegółów
        const detailsPromises = filteredItems.map((item) =>
          fetch(`/v1/api/${currentType}/id/${item.id}`, {
            signal: controller.signal,
          })
            .then((res) => (res.ok || res.status === 302 ? res.json() : null))
            .catch((err) => {
              if (err.name !== "AbortError") console.error(err);
              return null;
            })
        );

        const detailsResults = await Promise.all(detailsPromises);
        setAllLiked(detailsResults.filter((item) => item !== null));
      } catch (err) {
        if (err.name !== "AbortError") {
          console.error(err);
          setError("Błąd pobierania danych.");
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }

    fetchData();
    return () => controller.abort();
  }, [token, category]);

  async function handleUnlike(id) {
    const originalList = [...allLiked];
    setAllLiked((prev) =>
      prev.filter((item) => {
        const itemId = item.content.movie_id || item.content.id;
        return itemId !== id;
      })
    );

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
      if (!response.ok && response.status !== 302) throw new Error("Błąd");
    } catch (err) {
      setAllLiked(originalList);
      alert("Błąd usuwania.");
    }
  }

  return (
    <div className="max-w-6xl mx-auto mt-12 p-4 md:p-8 text-white min-h-[500px]">
      <h2 className="text-3xl font-bold mb-8 text-center text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-emerald-400">
        Twoja kolekcja
      </h2>

      <div className="flex justify-center mb-10">
        <CategorySwitch category={category} setCategory={setCategory} />
      </div>

      {loading && (
        <div className="flex flex-col justify-center items-center my-20 space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-4 border-slate-700 border-t-emerald-500"></div>
          <p className="text-slate-400 animate-pulse">Pobieranie...</p>
        </div>
      )}

      {!loading && !error && allLiked.length === 0 && (
        <div className="text-center py-20 bg-slate-800/30 rounded-2xl border-2 border-dashed border-slate-700">
          <div className="text-6xl mb-4 opacity-50">📭</div>
          <p className="text-xl text-slate-300 font-medium">Brak polubionych pozycji.</p>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {allLiked.map((item) => {
          const isBook = category === "ksiazki";
          const content = item.content;
          const itemId = isBook ? content.id : content.movie_id;
          const rating = isBook ? content.score : content.rating;

          // Rok wydania (dla książek często jest 0001, więc ukrywamy)
          let year = "---";
          if (content.release_date && !content.release_date.startsWith("0001")) {
            year = new Date(content.release_date).getFullYear();
          }

          if (!itemId) return null;

          return (
            <div
              key={itemId}
              className="group relative h-96 bg-slate-800 rounded-2xl overflow-hidden shadow-xl transition-transform duration-300 hover:-translate-y-2 border border-slate-700/50"
            >
              {/* --- WARSTWA 1: FRONT KARTY (WSPÓLNY) --- */}
              <div
                className={`absolute inset-0 bg-gradient-to-br ${
                  isBook ? "from-indigo-900 to-slate-900" : "from-emerald-900 to-slate-900"
                } flex flex-col items-center justify-center p-6 text-center`}
              >
                <div className="w-24 h-24 bg-white/10 rounded-full flex items-center justify-center mb-6 shadow-inner text-4xl backdrop-blur-sm">
                  {isBook ? "📖" : "🎬"}
                </div>

                <h3 className="text-xl font-bold text-white leading-tight line-clamp-2 px-2">
                  {content.title}
                </h3>

                <div className="mt-3 flex items-center gap-3 text-sm text-slate-300 font-mono">
                  {/* Rok pokazujemy tylko jeśli jest sensowny */}
                  {year !== "---" && <span>{year}</span>}
                  
                  {rating > 0 && (
                    <span className="bg-yellow-500/20 text-yellow-300 px-2 py-0.5 rounded border border-yellow-500/30">
                      ★ {rating}
                    </span>
                  )}
                </div>
              </div>

              {/* --- WARSTWA 2: HOVER (SZCZEGÓŁY) --- */}
              <div className="absolute inset-0 bg-slate-900/95 opacity-0 group-hover:opacity-100 transition-opacity duration-300 p-6 flex flex-col justify-between backdrop-blur-md">
                
                {/* Treść zależna od typu */}
                <div className="flex-1 overflow-hidden">
                  <div className="flex justify-between items-start mb-2">
                    <span
                      className={`text-[10px] font-bold uppercase tracking-widest px-2 py-1 rounded ${
                        isBook ? "bg-indigo-500/20 text-indigo-300" : "bg-emerald-500/20 text-emerald-300"
                      }`}
                    >
                      {isBook ? "KSIĄŻKA" : "TV SHOW"}
                    </span>
                  </div>

                  <h3 className="text-lg font-bold text-white mb-3 line-clamp-2 leading-snug">
                    {content.title}
                  </h3>

                  {isBook ? (
                    // --- WYGLĄD DLA KSIĄŻKI (METRYCZKA) ---
                    <div className="space-y-3">
                      {/* Autor */}
                      <div className="pb-2 border-b border-slate-700">
                        <span className="text-[10px] uppercase text-slate-500 font-bold block mb-1">Autor</span>
                        <p className="text-sm text-indigo-200 font-medium">{content.authors || "Autor nieznany"}</p>
                      </div>

                      {/* Siatka danych */}
                      <div className="grid grid-cols-2 gap-x-2 gap-y-3">
                        <div>
                           <span className="text-[10px] uppercase text-slate-500 font-bold block">Wydawca</span>
                           <p className="text-xs text-slate-300 truncate" title={content.publisher}>{content.publisher || "-"}</p>
                        </div>
                        <div>
                           <span className="text-[10px] uppercase text-slate-500 font-bold block">Strony</span>
                           <p className="text-xs text-slate-300">{content.pages || "?"}</p>
                        </div>
                        <div>
                           <span className="text-[10px] uppercase text-slate-500 font-bold block">Język</span>
                           <p className="text-xs text-slate-300 uppercase">{content.language || "-"}</p>
                        </div>
                        <div>
                           <span className="text-[10px] uppercase text-slate-500 font-bold block">Liczba ocen</span>
                           <p className="text-xs text-slate-300">{content.total_rating || 0}</p>
                        </div>
                      </div>
                      
                      {/* ISBN na dole */}
                      <div className="mt-2 pt-2 border-t border-slate-700/50">
                        <span className="text-[9px] text-slate-500 font-mono">ISBN: {content.isbn13 || content.isbn || "-"}</span>
                      </div>
                    </div>
                  ) : (
                    // --- WYGLĄD DLA FILMU (STARY) ---
                    <>
                      <div className="flex flex-wrap gap-2 mb-4">
                        {content.genres?.slice(0, 3).map((g) => (
                          <span
                            key={g.id}
                            className="text-[10px] bg-slate-800 text-slate-400 px-2 py-1 rounded border border-slate-700"
                          >
                            {g.name}
                          </span>
                        ))}
                      </div>
                      <p className="text-slate-400 text-xs leading-relaxed line-clamp-5 border-l-2 border-slate-700 pl-3 italic">
                        {content.overview || "Brak opisu."}
                      </p>
                    </>
                  )}
                </div>

                {/* Przycisk usuwania (wspólny) */}
                <button
                  onClick={() => handleUnlike(itemId)}
                  className="w-full mt-4 py-3 rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-semibold transition-colors flex items-center justify-center gap-2 shadow-lg shadow-rose-900/50"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path
                      fillRule="evenodd"
                      d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Usuń
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}