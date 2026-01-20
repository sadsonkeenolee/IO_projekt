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
        const details = await fetch(`/v1/api/tv/id/${item.id}`).catch((err) => {console.log(err); return null;});
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

      <div className="space-y-4 mt-6">
        {!loading && allLiked.length === 0 && (
          <p className="text-center text-slate-400">Brak pozycji w tej kategorii.</p>
        )}

        {allLiked.map((item) => (
          <div key={item.content.movie_id} className="bg-slate-700 p-4 rounded-lg flex justify-between items-center hover:bg-slate-600 transition-colors">
            <div>
              <p className="text-white font-semibold">{item.content.title}</p>
              <p className="text-slate-400 text-sm">{item.content.release_date}</p>
            </div>
            <button
              onClick={() => handleUnlike(item.content.movie_id)}
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
