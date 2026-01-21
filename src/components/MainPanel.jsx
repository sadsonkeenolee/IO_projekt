import { useState, useEffect } from "react";

export default function MainPanel({ category }) {
  const colors = {
    "filmy i seriale": "bg-emerald-900",
    ksiazki: "bg-indigo-900",
  };

  const [query, setQuery] = useState("");
  const [results, setResults] = useState([]);
  const [liked, setLiked] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const timeout = setTimeout(() => {
      if (!query) {
        setResults([]);
        return;
      }
      fetchMovies(query);
    }, 300);

    return () => clearTimeout(timeout);
  }, [query, category]);

  async function fetchMovies(searchQuery) {
    setLoading(true);
    setError("");
    setResults([]);

    try {
      let url = "";
      // Dynamiczne ustawianie endpointu w zależności od kategorii
      if (category === "filmy i seriale") {
        url = `/v1/api/tv/title/${searchQuery}`;
      } else if (category === "ksiazki") {
        url = `/v1/api/book/title/${searchQuery}`;
      }

      const res = await fetch(url);
      
      if (res.status === 404) {
        throw new Error("Nie znaleziono pozycji o tym tytule.");
      }
      
      if (!res.ok && res.status !== 302) throw new Error("Błąd połączenia z serwerem");

      const data = await res.json();
      
      // Zakładam, że struktura odpowiedzi dla książek jest podobna do TV (content.movie_id lub content.id)
      setResults([data]);
      console.log(data);
    } catch (err) {
      setResults([]);
      setError(err.message || "Nie udało się pobrać danych z serwera.");
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

    const typeMap = {
      "filmy i seriale": "tv",
      "ksiazki": "book",
    };

    try {
      const response = await fetch("/v1/auth/event/push", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          access_token: token,
          event: eventType, 
          type: typeMap[category],
          id: id.toString(),
        }),
      });

      if (!response.ok && response.status !== 302) {
        throw new Error("Błąd podczas komunikacji z serwerem");
      }
      
      setLiked((prev) =>
        isCurrentlyLiked 
          ? prev.filter((item) => item !== id) 
          : [...prev, id]                   
      );
    } catch (err) {
      console.error(`Błąd podczas ${eventType}:`, err);
      alert("Nie udało się zaktualizować statusu. Spróbuj ponownie.");
    }
  }

  // Helper do pobierania ID niezależnie od tego czy to film (movie_id) czy książka (id)
  const getItemId = (item) => item.content.movie_id || item.content.id;

  return (
    <div className={`${colors[category]} shadow-xl rounded-xl p-10 max-w-4xl mx-auto transition-colors duration-500 text-white`}>
      <div className="mb-12 text-center">
        <h2 className="text-2xl font-bold mb-6">
          {category === "filmy i seriale" && "Wyszukaj film lub serial, który lubisz"}
          {category === "ksiazki" && "Wyszukaj książkę, którą lubisz"}
        </h2>

        <input
          type="text"
          className="w-full px-4 py-3 rounded-lg bg-slate-700 border border-neutral-600 focus:outline-none focus:ring focus:ring-blue-500 mb-6 text-white"
          placeholder={category === "ksiazki" ? "np. Wiedźmin, Harry Potter..." : "np. Interstellar, Breaking Bad..."}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />

        {loading && <p>Wczytywanie...</p>}
        {error && <p className="text-yellow-400">{error}</p>}
      </div>

      <hr className="border-neutral-600 mb-12" />
      <div className="space-y-4">
        {/* {results.map((item) => { */}
        {/*   const id = getItemId(item); */}
        {/*   return ( */}
        {/*     <div key={id} className="bg-slate-700 p-4 rounded-lg flex justify-between items-center"> */}
        {/*       <div> */}
        {/*         <p className="font-semibold">{item.content.title || item.content.name}</p> */}
        {/*         <p className="text-slate-400 text-sm"> */}
        {/*           {item.content.release_date || item.content.published_date || "Data nieznana"} */}
        {/*         </p> */}
        {/*       </div> */}
        {/*       <button */}
        {/*         onClick={() => toggleLike(id)} */}
        {/*         className={`px-4 py-2 rounded-md font-medium transition-colors ${ */}
        {/*           liked.includes(id) */}
        {/*             ? "bg-rose-600 hover:bg-rose-700" */}
        {/*             : "bg-slate-500 hover:bg-slate-600" */}
        {/*         }`} */}
        {/*       > */}
        {/*         {liked.includes(id) ? "Lubisz to ❤️" : "Lubię 👍"} */}
        {/*       </button> */}
        {/*     </div> */}
        {/*   ); */}
        {/* })} */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {results.map((item) => (
          <div 
            key={item.content.movie_id} 
            className="group relative h-100 bg-slate-800 rounded-xl overflow-hidden shadow-lg transition-transform duration-300 hover:-translate-y-2"
          >
            <div className="absolute inset-0 w-full h-full bg-gradient-to-br from-slate-700 to-slate-900 flex flex-col items-center justify-center text-slate-500">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z" />
              </svg>
              <span className="font-bold text-lg uppercase tracking-wider">{item.content.title}</span>
            </div>

            <div className="absolute inset-0 bg-slate-900/90 opacity-0 group-hover:opacity-100 transition-opacity duration-300 p-6 flex flex-col justify-between">
              <div>
                <div className="flex justify-between items-start">
                  <h3 className="text-xl font-bold text-white leading-tight">{item.content.title}</h3>
                  <span className="bg-yellow-500 text-slate-900 text-xs font-bold px-2 py-1 rounded">
                    ★ {item.content.rating}
                  </span>
                </div>

                <p className="text-slate-400 text-xs mt-1">
                  {new Date(item.content.release_date).getFullYear()} • {item.content.runtime} min
                </p>

                <div className="flex flex-wrap gap-1 mt-3">
                  {item.content.genres?.slice(0, 3).map(g => (
                    <span key={g.id} className="text-[10px] bg-slate-700 text-slate-300 px-2 py-0.5 rounded-full">
                      {g.name}
                    </span>
                  ))}
                </div>

                <p className="text-slate-300 text-sm mt-4 line-clamp-4 italic">
                  "{item.content.overview}"
                </p>
              </div>

              <button
                onClick={() => toggleLike(item.content.movie_id)}
                className={`w-full py-2.5 rounded-lg font-semibold transition-all flex items-center justify-center gap-2 ${
liked.includes(item.content.movie_id)
? "bg-rose-600 hover:bg-rose-700 text-white"
: "bg-white hover:bg-slate-200 text-slate-900"
}`}
              >
                {liked.includes(item.content.movie_id) ? (
                  <><span className="text-lg">❤️</span> Lubisz to</>
                ) : (

                    <><span className="text-lg">👍</span> Lubię</>
                  )}
              </button>
            </div>
          </div>
        ))}
      </div>

    </div>
    </div>



  );
}
