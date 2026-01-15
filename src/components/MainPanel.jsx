import { useState, useEffect } from "react";

export default function MainPanel({ category }) {
  const colors = {
    film: "bg-purple-900",
    serial: "bg-emerald-900",
    ksiazka: "bg-indigo-900",
  };

  const [query, setQuery] = useState("");
  const [results, setResults] = useState([]);
  const [liked, setLiked] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // debounce dla wyszukiwania live
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
      if (category === "film") url = `http://localhost:9997/v1/api/tv/title/${searchQuery}`;

      const res = await fetch(url);
      if (!res.ok) throw new Error("Brak połączenia z serwerem");

      const data = await res.json();
      setResults(data);
    } catch (err) {
      setResults([]);
      setError("Nie udało się pobrać danych z serwera.");
    }

    setLoading(false);
  }

  function toggleLike(id) {
    setLiked((prev) =>
      prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id]
    );
  }

  return (
    <div className={`${colors[category]} shadow-xl rounded-xl p-10 max-w-4xl mx-auto transition-colors duration-500`}>
      <div className="mb-12 text-center">
        <h2 className="text-2xl font-bold mb-6">
          {category === "film" && "Wyszukaj film, który lubisz"}
          {category === "serial" && "Wyszukaj serial, który lubisz"}
          {category === "ksiazka" && "Wyszukaj książkę, którą lubisz"}
        </h2>

        <input
          type="text"
          className="w-full px-4 py-3 rounded-lg bg-slate-700 border border-slate-600 focus:outline-none focus:ring focus:ring-blue-500 mb-6"
          placeholder="np. Interstellar, Breaking Bad, Avatar..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />

        {loading && <p className="text-white">Wczytywanie...</p>}
        {error && <p className="text-yellow-400">{error}</p>}
      </div>

      <hr className="border-slate-600 mb-12" />

      <div className="space-y-4">
        {results.map((item) => (
          <div key={item.id} className="bg-slate-700 p-4 rounded-lg flex justify-between items-center">
            <div>
              <p className="text-white font-semibold">{item.title}</p>
              {item.year && <p className="text-slate-400 text-sm">{item.year}</p>}
            </div>
            <button
              onClick={() => toggleLike(item.id)}
              className={`px-3 py-1 rounded-md font-medium ${
                liked.includes(item.id) ? "bg-green-600 text-white" : "bg-slate-500 text-white"
              }`}
            >
              {liked.includes(item.id) ? "Lubisz to ❤️" : "Lubię 👍"}
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
