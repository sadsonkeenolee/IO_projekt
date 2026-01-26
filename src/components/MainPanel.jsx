import { useState, useEffect } from "react";

export default function MainPanel({ category }) {
  const colors = {
    "filmy i seriale": "bg-emerald-900",
    ksiazki: "bg-indigo-900",
  };

  const [query, setQuery] = useState("");
  const [result, setResult] = useState(null); // Zmienione z tablicy [] na null
  const [liked, setLiked] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const timeout = setTimeout(() => {
      if (!query) {
        setResult(null);
        return;
      }
      fetchItem(query);
    }, 500); // Zwiększyłem lekko delay, żeby nie strzelać przy szybkim pisaniu

    return () => clearTimeout(timeout);
  }, [query, category]);

  async function fetchItem(searchQuery) {
    setLoading(true);
    setError("");
    setResult(null);

    try {
      let url = "";
      if (category === "filmy i seriale") {
        url = `/v1/api/tv/title/${searchQuery}`;
      } else if (category === "ksiazki") {
        url = `/v1/api/book/title/${searchQuery}`;
      }

      const res = await fetch(url);

      if (res.status === 404) {
        throw new Error("Nie znaleziono pozycji o tym tytule.");
      }

      if (!res.ok && res.status !== 302)
        throw new Error("Błąd połączenia z serwerem");

      const data = await res.json();
      setResult(data); // Ustawiamy obiekt, nie tablicę
    } catch (err) {
      setResult(null);
      setError(err.message || "Nie udało się pobrać danych.");
    } finally {
      setLoading(false);
    }
  }

  async function toggleLike(id) {
    const token = localStorage.getItem("token");

    if (!token) {
      setError("Musisz być zalogowany, aby wykonać tę akcję.");
      window.scrollTo({ top: 0, behavior: "smooth" });
      return;
    }

    const isCurrentlyLiked = liked.includes(id);
    const eventType = isCurrentlyLiked ? "dislike" : "like";
    const typeMap = { "filmy i seriale": "tv", ksiazki: "book" };

    try {
      const response = await fetch("/v1/auth/event/push", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          access_token: token,
          event: eventType,
          type: typeMap[category],
          id: id.toString(),
        }),
      });

      if (!response.ok && response.status !== 302) {
        throw new Error("Błąd serwera");
      }

      setLiked((prev) =>
        isCurrentlyLiked ? prev.filter((item) => item !== id) : [...prev, id]
      );
    } catch (err) {
      console.error(err);
      alert("Błąd aktualizacji statusu.");
    }
  }

  // --- RENDEROWANIE WYNIKU ---
  const renderSingleResult = () => {
    if (!result) return null;

    const isBook = category === "ksiazki";
    const content = result.content;
    const itemId = isBook ? content.id : content.movie_id;
    const ratingValue = isBook ? content.score : content.rating;
    
    // Formatowanie daty
    let year = "Rok nieznany";
    if (content.release_date && !content.release_date.startsWith("0001")) {
       year = new Date(content.release_date).getFullYear();
    } else if (isBook) {
        // Czasem książki mają rok w innym polu lub brak, tu fallback
        year = "---";
    }

    return (
      <div className="animate-fade-in-up bg-slate-800/80 backdrop-blur-sm rounded-2xl overflow-hidden shadow-2xl border border-slate-700">
        <div className="flex flex-col md:flex-row">
          
          {/* LEWA STRONA: IKONA / OCENA */}
          <div className="md:w-1/3 bg-slate-900/50 p-8 flex flex-col items-center justify-center border-b md:border-b-0 md:border-r border-slate-700">
             <div className="w-32 h-32 md:w-40 md:h-40 bg-slate-700 rounded-full flex items-center justify-center mb-6 shadow-inner text-slate-400">
                {isBook ? (
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-20 w-20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                  </svg>
                ) : (
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-20 w-20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z" />
                  </svg>
                )}
             </div>
             
             {ratingValue > 0 && (
               <div className="text-center">
                 <div className="text-4xl font-bold text-yellow-400 mb-1">{ratingValue}</div>
                 <div className="text-xs uppercase tracking-widest text-slate-500">Ocena</div>
                 {isBook && content.total_rating && (
                   <div className="text-xs text-slate-600 mt-1">({content.total_rating} głosów)</div>
                 )}
               </div>
             )}
          </div>

          {/* PRAWA STRONA: SZCZEGÓŁY */}
          <div className="md:w-2/3 p-8 flex flex-col justify-between">
            <div>
              <div className="flex justify-between items-start mb-4">
                <h2 className="text-3xl font-bold text-white leading-tight">{content.title}</h2>
                <span className="bg-slate-700 text-slate-300 px-3 py-1 rounded text-sm font-mono">
                  {year}
                </span>
              </div>

              {/* SEKCJONOWANIE DANYCH */}
              <div className="space-y-4 mb-8">
                
                {/* WARIANT KSIĄŻKA */}
                {isBook && (
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div className="bg-slate-700/30 p-3 rounded">
                      <span className="block text-slate-400 text-xs mb-1">Autor</span>
                      <span className="font-semibold">{content.authors || "Nieznany"}</span>
                    </div>
                    <div className="bg-slate-700/30 p-3 rounded">
                      <span className="block text-slate-400 text-xs mb-1">Wydawca</span>
                      <span className="font-semibold">{content.publisher || "Brak danych"}</span>
                    </div>
                    <div className="bg-slate-700/30 p-3 rounded">
                      <span className="block text-slate-400 text-xs mb-1">Liczba stron</span>
                      <span className="font-semibold">{content.pages}</span>
                    </div>
                    <div className="bg-slate-700/30 p-3 rounded">
                      <span className="block text-slate-400 text-xs mb-1">ISBN</span>
                      <span className="font-mono text-xs md:text-sm">{content.isbn || content.isbn13 || "-"}</span>
                    </div>
                  </div>
                )}

                {/* WARIANT TV/FILM */}
                {!isBook && (
                  <>
                    <div className="flex flex-wrap gap-2 mb-4">
                      {content.genres?.map((g) => (
                        <span key={g.id} className="text-xs bg-indigo-600 text-white px-3 py-1 rounded-full">
                          {g.name}
                        </span>
                      ))}
                      <span className="text-xs bg-slate-700 text-slate-300 px-3 py-1 rounded-full border border-slate-600">
                        {content.runtime} min
                      </span>
                    </div>
                    <p className="text-slate-300 leading-relaxed italic border-l-4 border-indigo-500 pl-4 py-1">
                      "{content.overview || "Brak opisu."}"
                    </p>
                  </>
                )}
              </div>
            </div>

            {/* PRZYCISK AKCJI */}
            <button
              onClick={() => toggleLike(itemId)}
              className={`w-full py-4 rounded-xl font-bold text-lg transition-all flex items-center justify-center gap-3 shadow-lg ${
                liked.includes(itemId)
                  ? "bg-rose-600 hover:bg-rose-700 text-white shadow-rose-900/50"
                  : "bg-white hover:bg-slate-200 text-slate-900 shadow-white/10"
              }`}
            >
              {liked.includes(itemId) ? (
                <>
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z" clipRule="evenodd" />
                  </svg>
                  <span>Dodano do ulubionych</span>
                </>
              ) : (
                <>
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
                  </svg>
                  <span>Lubię to!</span>
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className={`${colors[category]} min-h-[500px] shadow-2xl rounded-3xl p-6 md:p-12 max-w-5xl mx-auto transition-colors duration-500 text-white`}>
      <div className="max-w-2xl mx-auto text-center mb-10">
        <h2 className="text-3xl md:text-4xl font-extrabold mb-8 tracking-tight">
          {category === "filmy i seriale" ? "Znajdź film lub serial" : "Znajdź książkę"}
        </h2>

        <div className="relative group">
           <div className="absolute -inset-1 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-lg blur opacity-25 group-hover:opacity-75 transition duration-1000 group-hover:duration-200"></div>
           <input
            type="text"
            className="relative w-full px-6 py-4 text-lg rounded-lg bg-slate-800 border border-slate-600 focus:outline-none focus:border-indigo-500 text-white placeholder-slate-400 shadow-xl"
            placeholder={category === "ksiazki" ? "Wpisz tytuł książki..." : "Wpisz tytuł filmu..."}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>

        {loading && <p className="mt-4 text-slate-300 animate-pulse">Przeszukiwanie bazy...</p>}
        {error && <p className="mt-4 text-rose-400 bg-rose-900/20 py-2 px-4 rounded-lg inline-block">{error}</p>}
      </div>

      {renderSingleResult()}
      
      {!result && !loading && !error && (
         <div className="text-center text-slate-400/50 mt-12">
            <p className="text-sm uppercase tracking-widest">Brak wyników do wyświetlenia</p>
         </div>
      )}
    </div>
  );
}