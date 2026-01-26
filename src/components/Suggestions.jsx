import { useEffect, useState } from "react";

export default function Suggestions({ category = "filmy i seriale" }) {
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(true);

  // Funkcja pomocnicza do kolorów w zależności od typu
  const getTypeStyles = (type) => {
    if (type === "book") {
      return {
        border: "border-indigo-500/30 hover:border-indigo-400",
        bg: "bg-gradient-to-br from-slate-800 to-indigo-950/20",
        text: "text-indigo-300",
        icon: (
          <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
          </svg>
        ),
        label: "Książka"
      };
    } else {
      return {
        border: "border-emerald-500/30 hover:border-emerald-400",
        bg: "bg-gradient-to-br from-slate-800 to-emerald-950/20",
        text: "text-emerald-300",
        icon: (
          <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z" />
          </svg>
        ),
        label: "Film / TV"
      };
    }
  };

  useEffect(() => {
    async function fetchRecommendations() {
      setLoading(true);
      const token = localStorage.getItem("token");
      if (!token) {
        setLoading(false);
        return;
      }

      try {
        // 1. Pobieranie polubień
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
          }),
        ]);

        const dataTv = await respTv.json();
        const dataBooks = await respBooks.json();

        const likedItemsForML = [
          ...(dataTv.content?.items || []).filter((i) => i.type === "tv").map((i) => ({ id: i.id, type: "movie" })),
          ...(dataBooks.content?.items || []).filter((i) => i.type === "book").map((i) => ({ id: i.id, type: "book" })),
        ];

        if (likedItemsForML.length === 0) {
          setLoading(false);
          return;
        }

        // 2. Pobranie rekomendacji
        const recResp = await fetch("/recommend", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            user_id: 1,
            liked_items: likedItemsForML,
            limit: 6,
          }),
        });

        if (recResp.ok) {
          const result = await recResp.json();
          setRecommendations(result.items || []);
        }
      } catch (error) {
        console.error(error);
      } finally {
        setLoading(false);
      }
    }

    fetchRecommendations();
  }, [category]);

  return (
    <div className="relative pt-10 pb-20 w-full max-w-[1400px] mx-auto px-6">
      
      {/* NAGŁÓWEK */}
      <div className="mb-10 border-l-4 border-slate-600 pl-6">
        <h2 className="text-3xl font-bold text-white mb-2">Sugerowane dla Ciebie</h2>
        <p className="text-slate-400">Na podstawie analizy Twoich ostatnich polubień</p>
      </div>

      {/* GRID KART */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {loading ? (
          <div className="col-span-full py-20 text-center text-slate-500 animate-pulse">
            Generowanie rekomendacji...
          </div>
        ) : recommendations.length > 0 ? (
          recommendations.map((item, idx) => {
            const styles = getTypeStyles(item.type);
            const matchPercent = Math.round(item.score * 100);

            return (
              <div
                key={`${item.type}-${item.id}-${idx}`}
                className={`relative group flex flex-col p-6 rounded-2xl border ${styles.border} ${styles.bg} shadow-xl transition-all duration-300 hover:shadow-2xl hover:-translate-y-1`}
              >
                {/* 1. HEADER KARTY: TYP I WYNIK */}
                <div className="flex justify-between items-start mb-4">
                  <div className={`flex items-center gap-2 px-3 py-1 rounded-full bg-slate-900/50 border border-slate-700/50 ${styles.text}`}>
                    {styles.icon}
                    <span className="text-xs font-bold uppercase tracking-wide">{styles.label}</span>
                  </div>
                  
                  <div className="text-right">
                    <div className="text-2xl font-black text-white">{matchPercent}%</div>
                    <div className="text-[10px] uppercase text-slate-500 font-bold tracking-wider">Dopasowanie</div>
                  </div>
                </div>

                {/* 2. TYTUŁ */}
                <h3 className="text-xl font-bold text-slate-100 leading-snug mb-4 group-hover:text-white transition-colors">
                  {item.title}
                </h3>

                {/* 3. POWÓD (REASON) */}
                <div className="mt-auto">
                   <div className="bg-slate-900/40 p-3 rounded-lg border border-slate-700/30">
                      <p className="text-xs text-slate-400 font-mono leading-relaxed">
                        <span className="text-slate-500 mr-2">ANALIZA:</span>
                        {item.reason}
                      </p>
                   </div>

                   {/* 4. GATUNKI (Jeśli są) */}
                   {item.genres && item.genres.length > 0 && (
                     <div className="flex flex-wrap gap-2 mt-4">
                       {item.genres.map((genre, gIdx) => (
                         <span key={gIdx} className="text-[10px] bg-slate-800 text-slate-300 px-2 py-1 rounded border border-slate-700">
                           {genre}
                         </span>
                       ))}
                     </div>
                   )}
                </div>

                {/* USUNIĘTO SEKCJĘ WATERMARKU TUTAJ */}

              </div>
            );
          })
        ) : (
          <div className="col-span-full py-16 text-center border-2 border-dashed border-slate-800 rounded-xl bg-slate-900/30">
            <p className="text-slate-400 text-lg">Brak wystarczających danych do analizy.</p>
            <p className="text-slate-600 text-sm mt-2">Polub więcej pozycji, aby otrzymać sugestie.</p>
          </div>
        )}
      </div>
    </div>
  );
}