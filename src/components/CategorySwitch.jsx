export default function CategorySwitch({ category, setCategory }) {
  const btnBase = "px-6 py-3 rounded-lg font-semibold shadow-md transition";

  return (
    <div className="flex justify-center gap-4 mb-10">

      <button
        onClick={() => setCategory("film")}
        className={`${btnBase} ${
          category === "film"
            ? "bg-purple-700 ring-4 ring-white"
            : "bg-purple-600 hover:bg-purple-700"
        }`}
      >
        🎬 Filmy
      </button>

      <button
        onClick={() => setCategory("serial")}
        className={`${btnBase} ${
          category === "serial"
            ? "bg-emerald-700 ring-4 ring-white"
            : "bg-emerald-600 hover:bg-emerald-700"
        }`}
      >
        📺 Seriale
      </button>

      <button
        onClick={() => setCategory("ksiazka")}
        className={`${btnBase} ${
          category === "ksiazka"
            ? "bg-indigo-700 ring-4 ring-white"
            : "bg-indigo-600 hover:bg-indigo-700"
        }`}
      >
        📚 Książki
      </button>

    </div>
  );
}