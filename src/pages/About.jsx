export default function About() {
  return (
    <div className="bg-slate-900 text-slate-200 p-6 sm:p-8 flex flex-col items-center">
      <div className="max-w-4xl w-full">
        
        {/* Nagłówek - lekko mniejszy margines dolny */}
        <header className="mb-8 border-b border-neutral-700 pb-6">
          <h1 className="text-3xl font-bold text-white mb-3">O Algorytmie Rekomendacji 🧠</h1>
          <p className="text-lg text-slate-400 font-light">
            System łączący analizę treści z danymi o zachowaniach użytkowników.
          </p>
        </header>

        {/* Odstępy gap-6 zamiast gap-8 (trochę ciasnej) */}
        <div className="grid gap-6">
          
          {/* Sekcja 1: Content-Based */}
          <section className="bg-slate-800/50 p-6 rounded-2xl border border-neutral-700">
            <h2 className="text-2xl font-semibold text-emerald-400 mb-3 flex items-center">
              Analiza treści (Content-based)
            </h2>
            <p className="leading-relaxed text-slate-300">
              Budujemy prosty indeks <strong>TF-IDF</strong>. Każdy przedmiot (film, książka, serial) reprezentowany jest jako wektor utworzony na podstawie słów występujących w tytule oraz przypisanych gatunków. Na tej podstawie, po polubieniu kilku pozycji, tworzony jest profil użytkownika będący sumą ich wektorów. Dopasowanie kandydatów odbywa się z wykorzystaniem <strong>podobieństwa cosinusowego</strong>.
            </p>
          </section>

          {/* Sekcja 2: Collaborative Filtering */}
          <section className="bg-slate-800/50 p-6 rounded-2xl border border-neutral-700">
            <h2 className="text-2xl font-semibold text-indigo-400 mb-3 flex items-center">
              Filtrowanie współpracujące (Collaborative filtering)
            </h2>
            <p className="leading-relaxed text-slate-300">
              Równolegle wykorzystywany jest moduł oparty na <strong>grafie interakcji</strong>. Analizuje on współwystępowanie polubień: jeśli osoby, które polubiły dany przedmiot, często mają polubione również inne konkretne pozycje, są one brane pod uwagę. W algorytmie stosowane jest <strong>logarytmiczne ważenie</strong>, które ogranicza wpływ bardzo aktywnych użytkowników.
            </p>
          </section>

          {/* Sekcja 3: MMR i Ranking */}
          <section className="bg-slate-800/50 p-6 rounded-2xl border border-neutral-700">
            <h2 className="text-2xl font-semibold text-white mb-3">Ranking i Dywersyfikacja</h2>
            <p className="leading-relaxed text-slate-300">
              Wyniki obu podejść są łączone w jeden ranking. Jeżeli dla danego przedmiotu dostępny jest sygnał społeczny, ma on nieco większy wpływ na końcową ocenę. Na etapie końcowym stosowana jest dywersyfikacja przy użyciu algorytmu <strong>MMR (Maximal Marginal Relevance)</strong>, który równoważy trafność rekomendacji z ich różnorodnością.
            </p>
          </section>

          {/* Stopka */}
          <footer className="mt-6 text-slate-500 text-sm italic text-center">
            W sytuacji braku danych o użytkowniku, system zwraca listę najpopularniejszych przedmiotów.
          </footer>
        </div>
      </div>
    </div>
  );
}
