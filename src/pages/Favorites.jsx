import CategorySwitch from "../components/CategorySwitch";
import { useState, useEffect } from "react";

export default function Favorites({ token }) {
  const [category, setCategory] = useState("film");
  const [allLiked, setAllLiked] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const typeMap = {
    "filmy i seriale": "tv",
    "ksiazki": "book",
  };

  useEffect(() => {
    const controller = new AbortController();
    async function fetchData() {
      if (!token)  {
        return;
      }

      setLoading(true);
      setAllLiked([])
      let resp = await fetch("/v1/auth/event/pull", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          access_token: token,
          event: "like",
          type: typeMap[category],
        }),
        signal: controller.signal
      }).catch((err) => {console.log(err); return null;});


      if (!(resp?.ok || resp?.status === 302)) {
        console.log("Zły status");
        return;
      }

      let data = await resp.json().catch((err) => {console.log(err); return null});
      if (data === null) {
        return;
      }

      let items = data.content.items;
      if (!items) {
        setAllLiked([]);
        setLoading(false);
        console.log("Brak polubionych rzeczy.");
        return;
      }

      let results = []
      for (const item of items) {
        const details = await fetch(`/v1/api/${typeMap[category]}/id/${item.id}`).catch((err) => {console.log(err); return null;});
        if (!(details?.ok || details?.status===302)) {
          continue
        }
        let content = await details.json().catch((err) => {console.log(err); return null;});
        if (content) {
          results.push(content)
        }
        // else {
        //   console.log("cos pustego")
        // }
      }
      setAllLiked(results);
      setLoading(false);
    }
    fetchData();
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
        setAllLiked((prev) => prev.filter((item) => item.content.movie_id !== id));
      }
    } catch (err) {
      console.error("Błąd usuwania:", err);
    }
  }

  return (
    <div className="max-w-4xl mx-auto mt-24 p-8 bg-slate-800 rounded-2xl shadow-xl text-white">
      <h2 className="text-2xl font-bold mb-6 text-center">Twoje polubione treści</h2>

      <CategorySwitch category={category} setCategory={setCategory} />

      {loading && <p className="text-center my-4 animate-pulse">Synchronizacja z bazą...</p>}
      {error && <p className="text-red-400 text-center my-4">{error}</p>}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8 mt-6">
        {!loading && allLiked.length === 0 && (
          <p className="text-center text-slate-400">Brak pozycji w tej kategorii.</p>
        )}

        {allLiked.map((item) => (
          <div 
            key={item.content.movie_id} 
            className="group relative h-100 bg-slate-800 rounded-2xl overflow-hidden shadow-2xl transition-all duration-500 hover:scale-[1.02]"
          >
            <div className="absolute inset-0 bg-gradient-to-b from-slate-700 to-slate-900 flex flex-col items-center justify-center p-6 text-center">
              <div className="w-20 h-20 bg-slate-800 rounded-full flex items-center justify-center mb-4 border border-slate-700 shadow-inner text-3xl">
                {category === "film" ? "🎬" : "📖"}
              </div>
              <h3 className="text-xl font-bold text-white/80 px-4">{item.content.title}</h3>
              <p className="text-slate-500 text-sm mt-2">
                {item.content.release_date ? new Date(item.content.release_date).getFullYear() : 'N/A'}
              </p>
            </div>

            <div className="absolute inset-0 bg-slate-900/95 opacity-0 group-hover:opacity-100 transition-all duration-300 p-8 flex flex-col">
              <div className="flex-1">
                <div className="flex justify-between items-start">
                  <span className="text-xs font-bold tracking-widest text-rose-500 uppercase">{typeMap[category]}</span>
                  <div className="bg-yellow-500 text-black text-[10px] font-black px-2 py-0.5 rounded">
                    ★ {item.content.rating || "N/A"}
                  </div>
                </div>

                <h3 className="text-2xl font-bold text-white mt-2 leading-tight">{item.content.title}</h3>

                <div className="flex flex-wrap gap-2 mt-4">
                  {item.content.genres?.slice(0, 3).map(g => (
                    <span key={g.id} className="text-[10px] border border-slate-700 text-slate-400 px-2 py-1 rounded-md">
                      {g.name}
                    </span>
                  ))}
                </div>

                <p className="text-slate-400 text-sm mt-6 line-clamp-6 italic leading-relaxed">
                  {item.content.overview || "Brak opisu dla tego tytułu."}
                </p>
              </div>

              <button
                onClick={() => handleUnlike(item.content.movie_id)}
                className="mt-4 w-full py-3 rounded-xl bg-red-500/10 hover:bg-red-600 text-red-500 hover:text-white border border-red-500/20 transition-all duration-300 font-bold flex items-center justify-center gap-2"
              >
                <span>💔</span> Odlub
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
