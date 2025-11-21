export default function MainPanel({ category }) {
  const colors = {
    film: "bg-purple-900",
    serial: "bg-emerald-900",
    ksiazka: "bg-indigo-900",
  };

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
          className="w-full px-4 py-3 rounded-lg bg-slate-700 border border-slate-600 
                     focus:outline-none focus:ring focus:ring-blue-500 mb-6"
          placeholder="np. Interstellar, Breaking Bad, Wiedźmin..."
        />

        <button className="px-8 py-3 bg-blue-600 hover:bg-blue-700 rounded-lg 
                           font-semibold text-lg shadow-lg">
          🔍 Pokaż podpowiedzi
        </button>
      </div>

      <hr className="border-slate-600 mb-12" />

    </div>
  );
}
