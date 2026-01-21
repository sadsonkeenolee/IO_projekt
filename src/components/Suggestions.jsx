import { useEffect, useState, useRef } from "react";

function AutoScrollingSection({ title, items, type, icon }) {
  const scrollRef = useRef(null);
  const [isPaused, setIsPaused] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [startX, setStartX] = useState(0);
  const [scrollLeft, setScrollLeft] = useState(0);

  useEffect(() => {
    const scrollContainer = scrollRef.current;
    if (!scrollContainer || isPaused || isDragging) return;

    const interval = setInterval(() => {
      const maxScrollLeft = scrollContainer.scrollWidth - scrollContainer.clientWidth;
      if (scrollContainer.scrollLeft >= maxScrollLeft - 1) {
        scrollContainer.scrollTo({ left: 0, behavior: "smooth" });
      } else {
        scrollContainer.scrollLeft += 1;
      }
    }, 40);

    return () => clearInterval(interval);
  }, [isPaused, isDragging]);

  const handleMouseDown = (e) => {
    setIsDragging(true);
    setStartX(e.pageX - scrollRef.current.offsetLeft);
    setScrollLeft(scrollRef.current.scrollLeft);
  };

  const stopDragging = () => setIsDragging(false);

  const handleMouseMove = (e) => {
    if (!isDragging) return;
    e.preventDefault();
    const x = e.pageX - scrollRef.current.offsetLeft;
    const walk = (x - startX) * 2;
    scrollRef.current.scrollLeft = scrollLeft - walk;
  };

  return (
    <section 
      onMouseEnter={() => setIsPaused(true)} 
      onMouseLeave={() => { setIsPaused(false); stopDragging(); }}
    >
      <div className="flex items-center gap-3 mb-8 px-2">
        <span className="text-3xl drop-shadow-md">{icon}</span>
        <h3 className="text-2xl font-black text-white uppercase tracking-tighter">{title}</h3>
      </div>

      <div 
        ref={scrollRef}
        onMouseDown={handleMouseDown}
        onMouseUp={stopDragging}
        onMouseMove={handleMouseMove}
        className="flex overflow-x-auto gap-6 pb-10 scrollbar-hide cursor-grab active:cursor-grabbing select-none"
        style={{ scrollBehavior: isDragging ? 'auto' : 'smooth' }}
      >
        {items.map((item) => (
          <div key={item.id || item.movie_id} className="flex-none w-72">
            <div className="group relative h-96 bg-slate-800 rounded-2xl overflow-hidden shadow-2xl transition-all duration-300 hover:-translate-y-2 border border-slate-700/50">

              <div className="absolute inset-0 bg-gradient-to-br from-slate-800 to-slate-950 flex flex-col items-center justify-center p-6 text-center">
                <div className="text-slate-600 mb-4 group-hover:scale-110 transition-transform duration-500">
                  {type === 'tv' ? (
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z" />
                    </svg>
                  ) : (
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                      </svg>
                    )}
                </div>
                <span className="font-black text-lg text-slate-400 uppercase tracking-widest leading-tight">
                  {item.title}
                </span>
              </div>

              <div className="absolute inset-0 bg-slate-950/95 opacity-0 group-hover:opacity-100 transition-all duration-300 p-8 flex flex-col justify-between backdrop-blur-sm">
                <div>
                  <div className="flex justify-between items-start mb-2">
                    <h3 className="text-xl font-bold text-white leading-tight">{item.title}</h3>
                    <span className="bg-yellow-500 text-slate-900 text-[10px] font-black px-2 py-1 rounded shadow-lg">
                      ★ {item.rating || item.score}
                    </span>
                  </div>

                  <p className="text-slate-400 text-xs font-medium">
                    {item.release_date ? new Date(item.release_date).getFullYear() : '2024'} 
                    {item.runtime ? ` • ${item.runtime} min` : ` • ${item.authors || 'Autor nieznany'}`}
                  </p>

                  <p className="text-slate-300 text-sm mt-6 line-clamp-6 italic leading-relaxed border-l-2 border-rose-500 pl-4">
                    "{item.overview || "Brak opisu dla tej pozycji."}"
                  </p>
                </div>

                <button className="w-full py-3 bg-white hover:bg-rose-600 hover:text-white text-slate-900 rounded-xl font-bold transition-all flex items-center justify-center gap-2 transform active:scale-95 shadow-xl">
                  <span className="text-lg">👍</span> Lubię to
                </button>
              </div>

            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

export default function Suggestions() {
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState({ books: [], shows: [] });

  async function fetchHome() {
    try {
      const response = await fetch("http://localhost:9997/v1/api/home/");
      const result = await response.json();
      setData({
        books: result.content.books || [],
        shows: result.content.shows || []
      });
    } catch (err) {
      console.error("Błąd pobierania:", err);
    } finally {
      setLoading(false);
    }
  }
  fetchHome()

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
    <div className="m-10 max-w-4xl mx-auto p-4">
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
      <AutoScrollingSection 
        title="Rekomendowane show" 
        items={[...data.shows]}
        type="tv" 
        icon="🎬" 
      />    
      <AutoScrollingSection 
        title="Rekomendowane książki" 
        items={[...data.books]}
        type="book" 
        icon="🎬" 
      />
    </div>
  );
}
