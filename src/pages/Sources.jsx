export default function DataSources() {
  return (
    <div className="bg-slate-900 text-slate-200 p-4 sm:p-6 flex flex-col items-center">
      {/* Zmieniono max-w-4xl na max-w-3xl, żeby był optycznie węższy, lub zostaw 4xl jeśli ma pasować szerokością do About */}
      <div className="max-w-4xl w-full">
        
        {/* Nagłówek - mniejsze marginesy (mb-6 zamiast mb-8) */}
        <header className="mb-6 border-b border-slate-700 pb-4">
          <h1 className="text-2xl font-bold text-white mb-2">Źródła Danych 📊</h1>
          <p className="text-base text-slate-400 font-light">
            Nasze modele trenujemy na publicznie dostępnych zbiorach danych.
          </p>
        </header>

        {/* Grid z mniejszym gapem (gap-4 zamiast gap-6) */}
        <div className="grid gap-4">
          
          {/* Sekcja 1: Książki */}
          <section className="bg-slate-800/50 p-5 rounded-xl border border-slate-700">
            <h2 className="text-xl font-semibold text-amber-400 mb-2 flex items-center">
              Książki (Goodreads) 📚
            </h2>
            <p className="text-sm leading-relaxed text-slate-300 mb-3">
              Moduł literacki korzysta z danych z serwisu Goodreads – zarówno grafu interakcji (UCSD), jak i metadanych (Kaggle).
            </p>
            <ul className="space-y-1.5 text-sm">
              <li>
                <a 
                  href="https://cseweb.ucsd.edu/~jmcauley/datasets/goodreads.html" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-blue-400 hover:text-blue-300 transition-colors flex items-center gap-2"
                >
                  🔗 UCSD Book Graph (J. McAuley)
                  <span className="text-slate-500 text-xs hidden sm:inline">- interakcje</span>
                </a>
              </li>
              <li>
                <a 
                  href="https://www.kaggle.com/datasets/jealousleopard/goodreadsbooks" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-blue-400 hover:text-blue-300 transition-colors flex items-center gap-2"
                >
                  🔗 Kaggle: Goodreads-books
                  <span className="text-slate-500 text-xs hidden sm:inline">- metadane</span>
                </a>
              </li>
              <li>
                <a 
                  href="https://www.kaggle.com/datasets/elvinrustam/books-dataset" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-blue-400 hover:text-blue-300 transition-colors flex items-center gap-2"
                >
                  🔗 Kaggle: Books Dataset
                  <span className="text-slate-500 text-xs hidden sm:inline">- uzupełnienie</span>
                </a>
              </li>
            </ul>
          </section>

          {/* Sekcja 2: Filmy */}
          <section className="bg-slate-800/50 p-5 rounded-xl border border-slate-700">
            <h2 className="text-xl font-semibold text-cyan-400 mb-2 flex items-center">
              Filmy (TMDB) 🎬
            </h2>
            <p className="text-sm leading-relaxed text-slate-300 mb-3">
              Rekomendacje filmowe opierają się na metadanych z The Movie Database (TMDB), w tym obsadzie, gatunkach i słowach kluczowych.
            </p>
            <ul className="space-y-1.5 text-sm">
              <li>
                <a 
                  href="https://www.kaggle.com/datasets/tmdb/tmdb-movie-metadata" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-blue-400 hover:text-blue-300 transition-colors flex items-center gap-2"
                >
                  🔗 TMDB Movie Metadata
                  <span className="text-slate-500 text-xs hidden sm:inline">- kredyty</span>
                </a>
              </li>
              <li>
                <a 
                  href="https://www.kaggle.com/datasets/rounakbanik/the-movies-dataset" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-blue-400 hover:text-blue-300 transition-colors flex items-center gap-2"
                >
                  🔗 The Movies Dataset
                  <span className="text-slate-500 text-xs hidden sm:inline">- archiwum</span>
                </a>
              </li>
            </ul>
          </section>

          {/* Sekcja 3: Info */}
          <section className="bg-slate-800/50 p-5 rounded-xl border border-slate-700">
            <h2 className="text-xl font-semibold text-white mb-2">Przetwarzanie danych</h2>
            <p className="text-sm leading-relaxed text-slate-300">
              Wszystkie zbiory poddano <strong>czyszczeniu i normalizacji</strong> (usuwanie duplikatów, unifikacja tytułów), aby umożliwić efektywne obliczanie podobieństwa.
            </p>
          </section>

          <footer className="mt-4 text-slate-600 text-xs italic text-center">
            Linki prowadzą do zewnętrznych repozytoriów.
          </footer>
        </div>
      </div>
    </div>
  );
}