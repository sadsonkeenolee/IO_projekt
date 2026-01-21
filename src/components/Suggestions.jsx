import { useEffect, useState } from "react";

export default function Suggestions() {
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchData() {
      setLoading(true);
      try {
        const token = localStorage.getItem("token");

        // 1. Pobieramy polubienia równolegle dla obu kategorii
        const [respTv, respBooks] = await Promise.all([
          fetch("/v1/auth/event/pull", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ access_token: token, event: "like", type: "tv" }),
          }),
          fetch("/v1/auth/event/pull", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ access_token: token, event: "like", type: "book" }),
          })
        ]);

        const dataTv = await respTv.json();
        const dataBooks = await respBooks.json();

        // Łączymy i filtrujemy dane dla modelu ML
        const likedItemsForML = [
          ...(dataTv.content?.items || [])
            .filter(item => item.type === "tv")
            .map(item => ({ id: item.id, type: "movie" })),
          ...(dataBooks.content?.items || [])
            .filter(item => item.type === "book")
            .map(item => ({ id: item.id, type: "book" }))
        ];

        if (likedItemsForML.length === 0) {
          setRecommendations([]);
          setLoading(false);
          return;
        }

        // 2. Zapytanie do serwisu ML (pamiętaj o pełnym adresie http://localhost:8001/recommend jeśli nie masz proxy)
        const recResp = await fetch("/recommend", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            user_id: 1,
            liked_items: likedItemsForML,
            limit: 10
          }),
        });

        const result = await recResp.json();

        // 3. ML zwraca obiekt z polem "items", który zawiera już tytuły
        // Przypisujemy te dane bezpośrednio do stanu
        setRecommendations(result.items || []);

      } catch (error) {
        console.error("Błąd w procesie rekomendacji:", error);
      } finally {
        setLoading(false);
      }
    }

    fetchData();
  }, []);

  if (loading) return <div className="text-center mt-10 text-white">Dobieranie rekomendacji...</div>;

  return (
    <div className="mt-6 max-w-4xl mx-auto p-4">
      <h2 className="text-2xl font-bold mb-6 text-center text-white">Twoje podpowiedzi</h2>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {recommendations.length > 0 ? (
          recommendations.map((item) => (
            <div key={`${item.type}-${item.id}`} className="bg-slate-700 p-5 rounded-xl shadow-md border border-neutral-600 flex flex-col justify-between">
              <div>
                <div className="flex justify-between items-start">
                  <h3 className="text-lg font-semibold text-white mb-1">{item.title}</h3>
                  <span className="text-[10px] bg-blue-600 text-white px-2 py-0.5 rounded uppercase">
                    {item.type}
                  </span>
                </div>
                
                {item.reason && (
                  <p className="text-xs text-blue-400 italic mb-3">{item.reason}</p>
                )}

                <div className="flex flex-wrap gap-1 mb-3">
                  {item.genres?.map((genre, idx) => (
                    <span key={idx} className="text-[10px] bg-slate-800 text-slate-300 px-2 py-0.5 rounded">
                      {genre}
                    </span>
                  ))}
                </div>
              </div>
              
              <div className="mt-2 pt-2 border-t border-slate-600 flex justify-between items-center">
                 <span className="text-xs text-slate-400 font-mono">Dopasowanie: {(item.score * 100).toFixed(0)}%</span>
                 <button className="text-xs bg-slate-600 hover:bg-slate-500 text-white px-3 py-1 rounded transition">
                   Szczegóły
                 </button>
              </div>
            </div>
          ))
        ) : (
          <div className="col-span-full text-center text-slate-400">
            Polub więcej filmów lub książek, aby otrzymać rekomendacje!
          </div>
        )}
      </div>
    </div>
  );
}