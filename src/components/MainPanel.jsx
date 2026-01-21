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
        {results.map((item) => {
          const id = getItemId(item);
          return (
            <div key={id} className="bg-slate-700 p-4 rounded-lg flex justify-between items-center">
              <div>
                <p className="font-semibold">{item.content.title || item.content.name}</p>
                <p className="text-slate-400 text-sm">
                  {item.content.release_date || item.content.published_date || "Data nieznana"}
                </p>
              </div>
              <button
                onClick={() => toggleLike(id)}
                className={`px-4 py-2 rounded-md font-medium transition-colors ${
                  liked.includes(id)
                    ? "bg-rose-600 hover:bg-rose-700"
                    : "bg-slate-500 hover:bg-slate-600"
                }`}
              >
                {liked.includes(id) ? "Lubisz to ❤️" : "Lubię 👍"}
              </button>
            </div>
          );
        })}
      </div>
    </div>
  );
}